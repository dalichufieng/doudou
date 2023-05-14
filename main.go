package main // 声明 main 包，表明当前是一个可执行程序

import (
	"encoding/json"
	"fmt"
	"goproject/config"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	uuid "github.com/satori/go.uuid"
) // 导入内置 fmt 包

type RequestFile struct {
	Queryurl      string `json:"queryurl"`
	Methodname    string `json:"methodname"`
	Requestbody   string `json:"Requestbody"`
	Requestheader http.Header
}

func param(res http.ResponseWriter, req *http.Request) {
	configInfo := config.InitConf()
	header := req.Header
	requestFile := RequestFile{}
	//赋值请求方法
	requestFile.Methodname = req.Method
	//获取请求地址
	requestUri := req.RequestURI
	//匹配域名即ipport.json需要映射的key
	compileRegex := regexp.MustCompile("^/[a-zA-Z0-9][-a-zA-Z0-9]{0,62}/")
	matchYu := compileRegex.FindStringSubmatch(requestUri)[0]
	if matchYu != "" {
		matchYu = strings.ReplaceAll(matchYu, "/", "")
	}
	//fmt.Println(matchYu)
	//获取映射的ip:port,组装真实的请求地址
	ipandport := ipport(matchYu)
	realrequestUri := strings.Replace(requestUri, "/"+matchYu, "", -1)
	requestFile.Queryurl = ipandport + realrequestUri
	//获取请求头
	requestFile.Requestheader = header
	//获取请求体
	var bodyBytes []byte
	if req.Body != nil {
		bodyBytes, _ = ioutil.ReadAll(req.Body)
		requestFile.Requestbody = string(bodyBytes)
	}
	//打印请求组装的数据
	// jsonBytes, err := json.Marshal(requestFile)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
	// personJSON := string(jsonBytes)
	// fmt.Println(personJSON)
	//fmt.Fprintln(res, req.Method)
	//检查4个交换文件夹是否存在
	_, isExitFileErr := os.Stat(configInfo.OsWrite)
	if os.IsNotExist(isExitFileErr) {
		fmt.Println("文件传夹不存在，创建文件夹")
		os.MkdirAll(configInfo.OsWrite, 0777)
	}
	_, isExitFileErr = os.Stat(configInfo.OsRead)
	if os.IsNotExist(isExitFileErr) {
		fmt.Println("文件传夹不存在，创建文件夹")
		os.MkdirAll(configInfo.OsRead, 0777)
	}
	_, isExitFileErr = os.Stat(configInfo.OsReqread)
	if os.IsNotExist(isExitFileErr) {
		fmt.Println("文件传夹不存在，创建文件夹")
		os.MkdirAll(configInfo.OsReqread, 0777)
	}
	_, isExitFileErr = os.Stat(configInfo.OsReqwrite)
	if os.IsNotExist(isExitFileErr) {
		fmt.Println("文件传夹不存在，创建文件夹")
		os.MkdirAll(configInfo.OsReqwrite, 0777)
	}
	// 将 JSON 写入本地文件
	fileName := GetUUID() + ".json"
	writeFileName := configInfo.OsWrite + "/" + fileName
	file, err := os.Create(writeFileName)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	encoder := json.NewEncoder(file)
	err = encoder.Encode(requestFile)
	if err != nil {
		panic(err)
	}
	//读取返回结果
	readFileName := configInfo.OsRead + "/" + fileName
	start := time.Now()
	var t time.Time
	for {
		_, errf := os.Stat(readFileName)
		if errf == nil {
			fmt.Println("File exist")
			content, _ := ioutil.ReadFile(readFileName)
			//json.NewEncoder(res).Encode(string(content))
			res.Write(content)
			//fmt.Fprintf(res, "%s", string(content))
			break
		}
		//每50毫秒检测一下文件是否存在
		time.Sleep(50 * time.Millisecond)
		t = time.Now()
		//fmt.Println(t.Sub(start).Milliseconds())
		if t.Sub(start).Milliseconds() > configInfo.RequestTimeOut {
			json.NewEncoder(res).Encode(string("请求超时"))
			break
		}
		fmt.Println("每50毫秒检测一下文件是否存在")
	}
	fmt.Println("timeout3")
	//json.NewEncoder(res).Encode(requestFile)
}

func ListenFolderNew(reqread string, reqwrite string) {
	fmt.Print("-----文件夹监听-------")
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()
	err = watcher.Add(reqread)
	done2 := make(chan bool)
	go func() {
		for {
			select {
			case event := <-watcher.Events:
				fmt.Println("***********event")
				//fmt.Println("event.Op=>%#v", event.Op)
				//fmt.Println("文件操作类型判断是不是新建一个文件：%#v", event.Op&fsnotify.Create == fsnotify.Create)
				if event.Op&fsnotify.Create == fsnotify.Create {
					fmt.Println("*Create**event")
					fmt.Println("新的文件:", event.Name)
					_, fileName := filepath.Split(event.Name)
					content, _ := ioutil.ReadFile(event.Name)
					var requestFile RequestFile
					err := json.Unmarshal([]byte(content), &requestFile)
					if err != nil {
						fmt.Println("error:", err)
						return
					}
					pl := strings.NewReader(requestFile.Requestbody)
					req, err := http.NewRequest(requestFile.Methodname, requestFile.Queryurl, pl)
					if err != nil {
						fmt.Println(err)
						return
					}
					//var result map[string]interface{}
					req.Header = requestFile.Requestheader
					client := &http.Client{}
					response, _ := client.Do(req)
					defer response.Body.Close()
					body, _ := ioutil.ReadAll(response.Body)
					// if err == nil {
					// 	err = json.Unmarshal(body, &result)
					// }
					file, err := os.Create(reqwrite + "/" + fileName)
					if err != nil {
						panic(err)
					}
					defer file.Close()
					//encoder := json.NewEncoder(file)
					//err = encoder.Encode(body)
					file.Write(body)
					if err != nil {
						panic(err)
					}

				}
			case err := <-watcher.Errors:
				log.Println("error:", err)
				// case <-time.After(60 * time.Second):
				// 	continue
			}

		}
	}()

	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("123")
	<-done2
	fmt.Println("456")
}

func GetUUID() string {
	id := uuid.NewV4()
	ids := id.String()
	return ids
}

func ipport(who string) (ip string) {
	r := ""
	data, err := ioutil.ReadFile("ipport.json")
	if nil != err {
		log.Fatalln("ReadFile ERROR:", err)
		//return
	} else {
		log.Println("ReadFile OK :\r\n", string(data))
	}
	var appConfig map[string]string
	err = json.Unmarshal(data, &appConfig)
	if nil != err {
		log.Fatalln("Unmarshal ERROR:", err)
		//return
	} else {
		for k, v := range appConfig {
			log.Println(k, " :", v)
			if who == k {
				r = v
				//fmt.Println(v)
			}
		}
	}
	return r
}

func main() { // main函数，是程序执行的入口
	configInfo := config.InitConf()
	fmt.Println("Hello World!") // 在终端打印 Hello World!
	go ListenFolderNew(configInfo.OsReqread, configInfo.OsReqwrite)
	server := http.Server{
		Addr: "localhost:" + configInfo.AppPort,
	}
	http.HandleFunc("/", param)
	server.ListenAndServe()

}
