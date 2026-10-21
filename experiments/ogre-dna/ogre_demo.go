package main

import (
 "fmt"
 "os"
 "strconv"
 "github.com/macawls/ogre/v3"
)
func main() {
 if len(os.Args)!=6 {panic("usage: dna-demo html width height font output")}
 input,err:=os.ReadFile(os.Args[1]);if err!=nil{panic(err)}
 width,err:=strconv.Atoi(os.Args[2]);if err!=nil{panic(err)}
 height,err:=strconv.Atoi(os.Args[3]);if err!=nil{panic(err)}
 fontData,err:=os.ReadFile(os.Args[4]);if err!=nil{panic(err)}
 renderer:=ogre.NewRenderer()
 fonts:=[]ogre.FontSource{}
 for _,name:=range []string{"DNAFont","DNAUnicode","Microsoft YaHei","sans-serif"}{fonts=append(fonts,ogre.FontSource{Name:name,Weight:400,Data:fontData},ogre.FontSource{Name:name,Weight:700,Data:fontData})}
 result,err:=renderer.Render(string(input),ogre.Options{Width:width,Height:height,Format:ogre.FormatPNG,Fonts:fonts,EmojiProvider:"none"})
 if err!=nil{panic(err)}
 if err:=os.WriteFile(os.Args[5],result.Data,0644);err!=nil{panic(err)}
 fmt.Println(len(result.Data))
}
