package main
import (
    "fmt"
    "os"
	"io"
	"net/http"
	"net/url"
	"encoding/json"
)
func main()  {
	http.HandleFunc("/", index)
	err := http.ListenAndServe(":8080", nil)  //监听8080端口
    if err != nil {
        fmt.Println("服务器开启错误:  %v", err)
    }
}

/*用户传入出各项参数*/
func index(w http.ResponseWriter, r *http.Request){   //设置索引页面
	//从请求中读取各项参数
	request_body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("读取请求数据错误: %v", err)
	}
	optin := 
	standard_option :=
	coord_type_input :=
	coord_type_output :=
	standard_track :=
	track :=
	// 轨迹重合率分析 API的地址
	API_host := "https://api.map.baidu.com"
	API_uri := "/trackmatch/v1/track"
	url := API_host + API_uri
	AK := os.Getenv("AK")   //从环境变量中获取AK值
	params := url.Values{
		"ak":                []string{AK},
		"option":            []string{option},
		"standard_option":   []string{standard_option},
		"coord_type_input":  []string{coord_type_input},
		"coord_type_output": []string{coord_type_output},
		"standard_track":    []string{standard_track},
		"track":             []string{track},
}
	//发送请求以调用API
	resp, err := http.PostForm(url, params)
	if err != nil {
	fmt.Printf("API服务器错误: %v", err)  //异常处理
	return
	}

	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)  //读取返回的数据
	if err != nil {
		fmt.Printf("读取响应数据错误: %v", err)  //异常处理
		return
	}
	//接下来解析得到的返回数据，并打印出"status" 和 "similarity"
	
		
}