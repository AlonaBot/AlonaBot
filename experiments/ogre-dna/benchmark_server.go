package main

import (
 "fmt"
 "log"
 "net/http"
 "os"

 ogre "github.com/macawls/ogre/v3"
)
func main(){
 if len(os.Args)!=5{log.Fatal("usage: server_bench <addr> <html> <font> <output>")}
 data,err:=os.ReadFile(os.Args[2]);if err!=nil{log.Fatal(err)}
 fontData,err:=os.ReadFile(os.Args[3]);if err!=nil{log.Fatal(err)}
 engine:=ogre.NewRenderer()
 for _,name:=range []string{"DNAFont","DNAUnicode","Microsoft YaHei","sans-serif"}{
   for _,w:=range []int{400,700}{
    if err:=engine.LoadFont(ogre.FontSource{Name:name,Weight:w,Data:fontData});err!=nil{log.Fatal(err)}
   }
 }
 htmlString:=string(data)
 http.HandleFunc("/render",func(w http.ResponseWriter,r *http.Request){
  image,err:=engine.Render(htmlString,ogre.Options{Width:1700,Height:900,Format:ogre.FormatPNG,EmojiProvider:"none"})
  if err!=nil{http.Error(w,err.Error(),500);return}
  w.Header().Set("Content-Type","image/png")
  _,_ = w.Write(image.Data)
 })
 fmt.Println("READY")
 log.Fatal(http.ListenAndServe(os.Args[1],nil))
}
