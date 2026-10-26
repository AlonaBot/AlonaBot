package packages

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/AlonaBot/AlonaBot/contracts/release"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
)

type archiveEntry struct { name string; kind byte; mode int64; body string }
func makeArchive(t *testing.T, entries []archiveEntry) []byte {
	t.Helper()
	var buf bytes.Buffer
	gz:=gzip.NewWriter(&buf);tw:=tar.NewWriter(gz)
	for _,e:=range entries {
		body:=[]byte(e.body)
		h:=&tar.Header{Name:e.name,Mode:e.mode,Typeflag:e.kind,Size:int64(len(body))}
		if e.kind==tar.TypeDir || e.kind==tar.TypeSymlink || e.kind==tar.TypeLink { h.Size=0 }
		if err:=tw.WriteHeader(h);err!=nil{t.Fatal(err)}
		if e.kind==tar.TypeReg || e.kind==tar.TypeRegA { if _,err:=tw.Write(body);err!=nil{t.Fatal(err)} }
	}
	if err:=tw.Close();err!=nil{t.Fatal(err)}
	if err:=gz.Close();err!=nil{t.Fatal(err)}
	return buf.Bytes()
}
func TestPreflightAndMoveTo(t *testing.T) {
	manifestYAML:=`manifest_version: 1
id: alonabot_plugin_demo
version: 1.2.3
protocol_min_major: 1
protocol_max_major: 1
required_capabilities: []
os: darwin
arch: arm64
entrypoint: bin/plugin
config_version: 1
source: github.com/example/demo
migration_mode: none
prepare_effects: none
`
	archive:=makeArchive(t,[]archiveEntry{{name:"plugin.yaml",kind:tar.TypeReg,mode:0o600,body:manifestYAML},{name:"bin/plugin",kind:tar.TypeReg,mode:0o755,body:"executable"}})
	pub,priv,err:=ed25519.GenerateKey(rand.Reader);if err!=nil{t.Fatal(err)}
	sum:=sha256.Sum256(archive)
	r:=release.Release{PublisherKeyID:release.KeyID(pub),Source:"github.com/example/demo",PluginID:"alonabot_plugin_demo",Version:"1.2.3",Assets:[]release.Asset{{OS:"darwin",Arch:"arm64",URL:"https://example.com/demo.tar.gz",Size:int64(len(archive)),SHA256:hex.EncodeToString(sum[:])}}}
	raw,err:=json.Marshal(r);if err!=nil{t.Fatal(err)}
	sig:=base64.StdEncoding.EncodeToString(ed25519.Sign(priv,raw))
	root:=t.TempDir();archivePath:=filepath.Join(root,"demo.tar.gz")
	if err:=os.WriteFile(archivePath,archive,0o600);err!=nil{t.Fatal(err)}
	stage:=filepath.Join(root,"staging")
	c,err:=Preflight(Input{ReleaseJSON:raw,Signature:[]byte(sig),PublicKey:pub,Source:r.Source,OS:"darwin",Arch:"arm64",ArchivePath:archivePath},stage,DefaultLimits)
	if err!=nil{t.Fatal(err)}
	if c.Manifest.ID!=r.PluginID{t.Fatalf("manifest = %#v",c.Manifest)}
	final,err:=c.MoveTo(filepath.Join(root,"packages"));if err!=nil{t.Fatal(err)}
	got,err:=os.ReadFile(filepath.Join(final,"bin","plugin"));if err!=nil||string(got)!="executable"{t.Fatalf("entrypoint=%q err=%v",got,err)}
	if _,err:=c.MoveTo(filepath.Join(root,"packages"));err==nil{t.Fatal("moved candidate twice")}
	if err:=c.Discard();err!=nil{t.Fatal(err)}
}
func TestPreflightRejectsSignedMetadataMismatch(t *testing.T) {
	archive:=makeArchive(t,[]archiveEntry{{name:"plugin.yaml",kind:tar.TypeReg,mode:0o600,body:"invalid"}})
	pub,priv,_:=ed25519.GenerateKey(rand.Reader);sum:=sha256.Sum256(archive)
	r:=release.Release{PublisherKeyID:release.KeyID(pub),Source:"github.com/example/demo",PluginID:"alonabot_plugin_demo",Version:"1.2.3",Assets:[]release.Asset{{OS:"darwin",Arch:"arm64",URL:"https://example.com/demo.tar.gz",Size:int64(len(archive)),SHA256:hex.EncodeToString(sum[:])}}}
	raw,_:=json.Marshal(r);sig:=base64.StdEncoding.EncodeToString(ed25519.Sign(priv,raw))
	root:=t.TempDir();archivePath:=filepath.Join(root,"demo.tar.gz");_ = os.WriteFile(archivePath,archive,0o600)
	_,err:=Preflight(Input{ReleaseJSON:raw,Signature:[]byte(sig),PublicKey:pub,Source:"github.com/other/demo",OS:"darwin",Arch:"arm64",ArchivePath:archivePath},filepath.Join(root,"stage"),DefaultLimits)
	var ve *validate.Error
	if !errors.As(err,&ve)||ve.Code!=pb.ErrorCode_ERROR_CODE_PERMISSION_DENIED{t.Fatalf("err=%v",err)}
}
func TestExtractRejectsUnsafeEntries(t *testing.T) {
	for _,tc:=range []struct{name string;entries []archiveEntry}{
		{name:"parent traversal",entries:[]archiveEntry{{name:"../x",kind:tar.TypeReg,mode:0o600,body:"x"}}},
		{name:"absolute",entries:[]archiveEntry{{name:"/abs",kind:tar.TypeReg,mode:0o600,body:"x"}}},
		{name:"backslash",entries:[]archiveEntry{{name:"a\\b",kind:tar.TypeReg,mode:0o600,body:"x"}}},
		{name:"symlink",entries:[]archiveEntry{{name:"link",kind:tar.TypeSymlink,mode:0o777}}},
		{name:"hardlink",entries:[]archiveEntry{{name:"link",kind:tar.TypeLink,mode:0o600}}},
		{name:"case collision",entries:[]archiveEntry{{name:"A.txt",kind:tar.TypeReg,mode:0o600,body:"a"},{name:"a.txt",kind:tar.TypeReg,mode:0o600,body:"b"}}},
	} { t.Run(tc.name,func(t *testing.T){archive:=makeArchive(t,tc.entries);dest:=filepath.Join(t.TempDir(),"out");if err:=os.Mkdir(dest,0o700);err!=nil{t.Fatal(err)};if err:=extractReader(bytes.NewReader(archive),dest,DefaultLimits);err==nil{t.Fatal("expected rejection")}}) }
}
