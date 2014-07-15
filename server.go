package main

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
)

type Sizer interface {
	Size() int64
}

var ServicePort string = ":8001"
var IPAddress string
var latest string

func DecodeStr(s string) int {
	src := []byte(s)
	dst := make([]byte, hex.DecodedLen(len(src)))
	hex.Decode(dst, src)

	return int(dst[0])
}

func Md5File(path string) string {

	file, inerr := os.Open(path)
	if inerr != nil {
		return ""
	}
	md5h := md5.New()
	io.Copy(md5h, file)
	return fmt.Sprintf("%x", md5h.Sum([]byte(""))) //md5
}

func ConvertToBin(File_Src, File_Target string) bool {
	var bdata int
	fmt.Println("-----Bin-----")
	if File_Src == "" || File_Target == "" {
		return false
	}
	if Md5File(File_Src) == "" {
		fmt.Println("Src file do not exist!")
		return false
	}
	FileInfo, _ := os.Stat(File_Src)
	fmt.Println("Src:", File_Src)
	fmt.Print("*FileSize:")
	fmt.Print(FileInfo.Size())
	fmt.Print("bytes ")
	fmt.Print("*MD5:")
	fmt.Println(Md5File(File_Src))

	fi, err := os.Open(File_Src)
	if err != nil {
		panic(err)
	}
	defer fi.Close()

	fo, err := os.Create(File_Target)
	if err != nil {
		panic(err)
	}
	defer fo.Close()

	br := bufio.NewReader(fi)
	for {
		line, err := br.ReadString('\n')
		if err == io.EOF {
			break
		} else {
			line = strings.Trim(line, " ")
			linelegal := line[0]
			if linelegal != ':' {
				return false
			}
			linesize := line[1:3]
			size := DecodeStr(linesize)
			linedatarec := line[7:9]
			isdatarec := DecodeStr(linedatarec)
			if isdatarec != 0 {
				continue
			}
			for h := 0; h < size; h++ {
				bdata = DecodeStr(line[9+h*2 : 9+h*2+2])
				buf := new(bytes.Buffer)
				binary.Write(buf, binary.LittleEndian, uint8(bdata))
				fo.Write(buf.Bytes())
			}
		}
	}
	fmt.Println("Target:", File_Target)
	return true
}

func HelloServer(w http.ResponseWriter, r *http.Request) {
	if "POST" == r.Method {
		file, _, err := r.FormFile("upl")
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		defer file.Close()
		f, err := os.Create("./public/upload/temp_file.tmp")
		defer f.Close()
		io.Copy(f, file)
		//fmt.Fprintln(w, "上传文件的大小为: ", file.(Sizer).Size())
		fi, err := os.Open("./public/upload/temp_file.tmp")
		if err != nil {
			panic(err)
		}
		defer fi.Close()
		md5h := md5.New()
		io.Copy(md5h, fi)
		File_Target := fmt.Sprintf("%x", md5h.Sum([]byte("")))
		ConvertSuccess := ConvertToBin("./public/upload/temp_file.tmp", "./public/upload/"+File_Target+".bin")
		if ConvertSuccess {
			fmt.Println("Operate success!")
			latest = File_Target
			w.Header().Add("Content-Type", "text/html")
			w.WriteHeader(200)
			html1 := `{"status":"success","target":"` + File_Target + `.bin"}`
			io.WriteString(w, html1)
		} else {
			fmt.Println("Operate failed!")
			latest = ""
			w.Header().Add("Content-Type", "text/html")
			w.WriteHeader(200)
			html := `{"status":"error"}`
			io.WriteString(w, html)
		}
		return
	}
}

func latestfile(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html")
	w.WriteHeader(200)
	//html := `{"target":"` + latest + `.bin"}`
	url := "/upload/" + latest + ".bin"
	html := `<script language="javascript">
    window.location = "` + url + `";` +`
	</script>`
	io.WriteString(w, html)
}

func illegal(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html")
	w.WriteHeader(200)
	html := "Original file illegal,please check your file!"
	io.WriteString(w, html)
}

func main() {
	conn, err1 := net.Dial("udp", "google.com:80")
	if err1 != nil {
		fmt.Println(err1.Error())
		return
	}
	defer conn.Close()
	IPAddress = strings.Split(conn.LocalAddr().String(), ":")[0]
	http.Handle("/", http.FileServer(http.Dir("public")))
	http.HandleFunc("/upload.php", HelloServer)
	http.HandleFunc("/latest", latestfile)
	http.HandleFunc("/upload/.bin", illegal)
	fmt.Println("Webserver serve at " + IPAddress + ServicePort)
	err := http.ListenAndServe(ServicePort, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
