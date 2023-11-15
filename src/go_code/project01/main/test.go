package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"net/url"
	"os"
	_ "github.com/go-sql-driver/mysql"
)

//设置罚时
const TP = 0.4
//获取AK
var AK = os.Getenv("AK")
var sql_host = os.Getenv("HOST")
var sql_password = os.Getenv("PASSWORD")

//设置记录历史搜索记录的结构体
type history struct{
	origin []string
	destination []string
}
//用于与数据库交互的引擎
type Engine struct{
	db *sql.DB
	table string `default:""` //表
	column []string `default:""`  //列
	value []string `default:""` //值
	where string `default:""`  //条件
	where_flag int 	//是否设置了条件
	   	
}

func StartupEngine(driver, source string) (e *Engine, err error) {
	db, err := sql.Open(driver, source)
	if err != nil {
		return
	}
	//发ping
	if err = db.Ping(); err != nil {
		return
	}
	e = new(Engine)
	e.db = db
	return e,nil
}
func (e *Engine) Where(_where string) *Engine{
	e.where_flag = 1
	e.where = _where	
	return e
}
func (e *Engine) Column(_column ...string) *Engine{
	e.column = append(e.column, _column...)
	return e
}
func (e *Engine) Value(value ...string) *Engine{
	e.value = append(e.value, value...)
	return e
}
func (e *Engine) Table(_table string) *Engine{
	e.table = _table
	return e
}

//定义查询方法
func (e *Engine) Select () *history {
	//构造查询语句
	cloumn_num := len(e.column)
	var column string
	var i int
	for i = 0; i < cloumn_num-1; i++ {
		column += e.column[i] + ","
	}
	column += e.column[i]
	var query string
	if (e.where_flag == 1){
		query = "select " + column +" from " + e.table + " where " + e.where
	}else{
		query = "select " + column +" from " + e.table
	}
	//进行查询
	rows,err := e.db.Query(query)
	if err != nil {
		fmt.Printf("查询错误：%v",err)
	}
	defer rows.Close()
	var h history  //用于记录查询所得数据
	var _origin string
	var _destination string
	//解析查询结果
	for rows.Next(){
		rows.Scan(&_origin,&_destination)
		h.origin = append(h.origin,_origin)
		h.destination = append(h.destination, _destination)
	}
	err = rows.Err()
	if err != nil {
		fmt.Printf("查询错误：%v",err)
	}
	return &h
}
//定义增加方法
func (e *Engine) Insert (){
	var columm string
	var value string
	cloumn_num := len(e.column)
	var i int
	for i = 0; i < cloumn_num - 1; i++ {
		columm += e.column[i] + ","
		value += "\"" + e.value[i] + "\"" + ","
	}
	columm += e.column[i]
	value += "\"" + e.value[i] + "\""
	query := "insert into " + e.table +" (" + columm +  ") " + "values " + "(" + value + ")"
	_,err := e.db.Exec(query)
	if err != nil{
		fmt.Printf("插入错误:%v",err)
	}
}
//定义删除方法
func (e *Engine) Delete() {
	var query string
	if e.where_flag == 1 {
		query = "delete from " + e.table + " where " + e.where
	}else{
		query = "delete from " + e.table
	}
	e.db.Exec(query)
}
func main()  {
	http.HandleFunc("/trackmatch",trackmatch)
	http.HandleFunc("/directionlite",directionlite)
	http.HandleFunc("/history",search_history)
	err := http.ListenAndServe(":8081", nil)  //监听8081端口
    if err != nil {
        fmt.Printf("服务器开启错误:  %v", err)
    }
}
/*用于将地址转换为经纬度*/
func geocoding(address string) string{
		API_host := "https://api.map.baidu.com"
		API_uri := "/geocoding/v3"
	
		// 设置负载
		params := url.Values {
			  "address": []string{address},
			  "output": []string{"json"},
			  "ak": []string{AK},
		}
	
		// 构造请求
		request, err := url.Parse(API_host + API_uri + "?" + params.Encode())
		if nil != err {
			fmt.Printf("请求构造错误: %v", err)
			return ""
		}
		//发起请求
		resp, err1 := http.Get(request.String())
		fmt.Printf("url: %s\n", request.String())
		defer resp.Body.Close()
		if err1 != nil {
			fmt.Printf("请求错误: %v", err1)
			return ""
		}
		body, err2 := io.ReadAll(resp.Body)
		if err2 != nil {
			fmt.Printf("读取相应信息错误: %v", err2)
		}
		//解析获取的数据
		//fmt.Printf(string(body)+"\n")

		type Location struct {
			Lng float64 `json:"lng"`
			Lat float64 `json:"lat"`
		}
		type Result struct {
			Location Location `json:"location"`
			Precise int `json:"precise"`
			Confidence int `json:"confidence"`
			Comprehension int `json:"comprehension"`
			Level string `json:"level"`
		}
		type Gcode struct {
			Status int `json:"status"`
			Result Result `json:"result"`
		}
		ResData := Gcode{}
		json.Unmarshal(body,&ResData)
		code := fmt.Sprintf("%f,%f",ResData.Result.Location.Lat,ResData.Result.Location.Lng)
		return code
}
/*用户以post方式传入各项参数*/
func trackmatch(w http.ResponseWriter, r *http.Request){
	
	//从请求中读取各项参数
	r.ParseForm()
	option := r.FormValue("option")
	standard_option := r.FormValue("standard_option")
	coord_type_input := r.FormValue("coord_type_input")
	coord_type_output := r.FormValue("coord_type_output")
	standard_track := r.FormValue("standard_track")
	track := r.FormValue("track")
	// 轨迹重合率分析 API的地址
	API_host := "https://api.map.baidu.com"
	API_uri := "/trackmatch/v1/track"


	// 设置请求参数
	params := url.Values{
		"ak":                []string{AK},
		"option":            []string{option},
		"standard_option":   []string{standard_option},
		"coord_type_input":  []string{coord_type_input},
		"coord_type_output": []string{coord_type_output},
		"standard_track":    []string{standard_track},
		"track":             []string{track},
	}
	
	// 发起请求
	target_url := API_host + API_uri
	resp, err := http.PostForm(target_url, params)
	if err != nil {
		fmt.Printf("API服务错误: %v", err)
		return
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("响应数据读取错误: %v", err)
		return
	}
	//接下来解析得到的返回数据，并打印出"status" 和 "similarity"
	//了解到可以自定义一个结构体，然后通过 Golang 的标准库 json 解析到定义的结构体中	
	type UtilsFunDataProcessed_standard_trackLoc struct {
		Longitude float64 `json:"longitude"`
		Latitude float64 `json:"latitude"`
	}
	type UtilsFunDataProcessed_standard_track struct {
		Loc UtilsFunDataProcessed_standard_trackLoc `json:"loc"`
		Loc_time int `json:"loc_time"`
	}
	
	type UtilsFunDataProcessed_trackLoc struct {
		Longitude float64 `json:"longitude"`
		Latitude float64 `json:"latitude"`
	}	

	type UtilsFunDataProcessed_track struct {
		Loc UtilsFunDataProcessed_trackLoc `json:"loc"`
		Loc_time int `json:"loc_time"`
		Unmatched int `json:"unmatched"`
	}

	type UtilsFunData struct {
		Similarity float64 `json:"similarity"`
		Processed_standard_track []UtilsFunDataProcessed_standard_track `json:"processed_standard_track"`
		Processed_track []UtilsFunDataProcessed_track `json:"processed_track"`
		Standard_track_distance float64 `json:"standard_track_distance"`
		Track_distance float64 `json:"track_distance"`
		Processed_standard_track_distance float64 `json:"processed_standard_track_distance"`
		Processed_track_distance float64 `json:"processed_track_distance"`
		Unmatched_distance float64 `json:"unmatched_distance"`
		Matched_distance float64 `json:"matched_distance"`
		Standard_match_ratio float64 `json:"standard_match_ratio"`
	}
	type UtilsFun struct {
		Status int `json:"status"`
		Message string `json:"message"`
		Data UtilsFunData `json:"data"`
	}
	//声明变量ResData用于存储获取的数据
	ResData := UtilsFun{}
	json.Unmarshal(body,&ResData)
	fmt.Fprintln(w,ResData.Status,ResData.Data.Similarity)
}
/*进行道路路况查询*/ 
func road(road_name string,city string) int{
	
	//API地址
	API_host := "https://api.map.baidu.com"
	API_uri := "/traffic/v1/road"
	//构建负载
	params := url.Values{
		"road_name": []string{road_name},
		"city": []string{city},
		"ak": []string{AK},
	}
	//构造请求
	request, err := url.Parse(API_host + API_uri + "?" + params.Encode())
    if nil != err {
        fmt.Printf("请求构造错误: %v", err)
        return -1
    }
	//发送请求
	resp,err1 := http.Get(request.String())
	if nil != err1{
		fmt.Printf("请求错误：%v",err1)
	}
	fmt.Printf("url: %s\n", request.String())
	defer resp.Body.Close()
	body,err2 := io.ReadAll(resp.Body)
	if nil != err2{
		fmt.Printf("解析响应信息错误：%v",err2)
	}
	//解析返回结果
	type Evaluation struct {
		Status int `json:"status"`
		StatusDesc string `json:"status_desc"`
	}
	type RoadTraffic struct {
		RoadName string `json:"road_name"`
	}
	type Road struct {
		Status int `json:"status"`
		Message string `json:"message"`
		Description string `json:"description"`
		Evaluation Evaluation `json:"evaluation"`
		RoadTraffic []RoadTraffic `json:"road_traffic"`
	}
	
	ResData := Road{}
	json.Unmarshal(body,&ResData)
	return ResData.Status  //返回状态码
}
/*用于查看历史记录*/
func search_history(w http.ResponseWriter, r *http.Request){
	fmt.Fprintf(w,"<h1>历史记录</h1>")
	//开启数据库
	dsn := fmt.Sprintf("%s:%s@tcp(127.0.0.1:3306)/web_server",sql_host,sql_password)
	e,errE := StartupEngine("mysql",dsn)
	if errE != nil{
		fmt.Fprintf(w,"错误:%v<br>",errE)
	}
	defer e.db.Close()
	h := new(history)
	h = e.Table("history").Column("origin","destination").Select()
	hlen := len(h.destination)
	for i := 0; i < hlen; i++ {
		fmt.Fprintf(w,"起点:%s 终点:%s<br>",h.origin[i],h.destination[i])
	}
}       
func directionlite(w http.ResponseWriter,r *http.Request){
	//设置页面用于接收用户的输入
	ht, err := template.ParseFiles("./directionlite.html")
	if err != nil {
	fmt.Fprintf(w, "解析页面错误: %v", err)
	return
	}
	ht.Execute(w,nil)
	//解析传入的参数
	r.ParseForm()
	transport := r.PostFormValue("transport")
	origin := r.PostFormValue("origin")
	destination := r.PostFormValue("destination")
	tactics := r.PostFormValue("tactics")
	mode := r.PostFormValue("mode")
	//存入历史记录
	if origin != "" && destination != ""{
		dsn := fmt.Sprintf("%s:%s@tcp(127.0.0.1:3306)/web_server",sql_host,sql_password)
		e,errE := StartupEngine("mysql",dsn)
		if errE != nil{
			fmt.Printf("错误:%v",errE)
		}
		defer e.db.Close()
		e.Table("history").Column("origin","destination").Value(origin,destination).Insert()  //插入该记录
	}
	//用地理编码解析地址
	origin  = geocoding(origin)
	destination = geocoding(destination)

	API_host := "https://api.map.baidu.com" 
	API_uri := "/directionlite/v1/" + transport
	
	// 设置请求参数
	params := url.Values {
		  "origin": []string{origin},
		  "destination": []string{destination},
		  "ak": []string{AK},
		  "tactics": []string{tactics},
	}
	
	// 发起请求
	request, err := url.Parse(API_host + API_uri + "?" + params.Encode())
	if nil != err {
		fmt.Printf("错误: %v", err)
		return
	}
	
	resp, err1 := http.Get(request.String())
	fmt.Printf("url: %s\n", request.String())
	defer resp.Body.Close()
	if err1 != nil {
		fmt.Printf("请求错误: %v", err1)
		return
	}
	body, err2 := io.ReadAll(resp.Body)
	if err2 != nil {
		fmt.Printf("读取响应信息错误: %v", err2)
	}
	//解析返回结果
	switch transport{
	case "driving":
		driving(w,body,mode)
	case "walking":
		walking(w,body,mode)
	case "riding":
		riding(w,body,mode)
	case "transit":
		transit(w,body,mode)
	}
	
}
func driving(w http.ResponseWriter,content []byte,mode string){
	type Origin struct {
		Lng float64 `json:"lng"`
		Lat float64 `json:"lat"`
	}
	type Destination struct {
		Lng float64 `json:"lng"`
		Lat float64 `json:"lat"`
	}
	type RestrictionInfo struct {
		Status int `json:"status"`
	}
	type TrafficCondition struct {
		Status int `json:"status"`
		GeoCnt int `json:"geo_cnt"`
	}
	type StartLocation struct {
		Lng string `json:"lng"`
		Lat string `json:"lat"`
	}
	type EndLocation struct {
		Lng string `json:"lng"`
		Lat string `json:"lat"`
	}
	type Steps struct {
		LegIndex int `json:"leg_index"`
		Distance int `json:"distance"`
		Duration int `json:"duration"`
		Direction int `json:"direction"`
		Turn int `json:"turn"`
		RoadType int `json:"road_type"`
		RoadTypes string `json:"road_types"`
		Instruction string `json:"instruction"`
		Path string `json:"path"`
		TrafficCondition []TrafficCondition `json:"traffic_condition"`
		StartLocation StartLocation `json:"start_location"`
		EndLocation EndLocation `json:"end_location"`
	}
	type Routes struct {
		RouteMd5 string `json:"route_md5"`
		Distance int `json:"distance"`
		Duration int `json:"duration"`
		TrafficCondition int `json:"traffic_condition"`
		Toll int `json:"toll"`
		RestrictionInfo RestrictionInfo `json:"restriction_info"`
		Steps []Steps `json:"steps"`
	}
	type Result struct {
		Origin Origin `json:"origin"`
		Destination Destination `json:"destination"`
		Routes []Routes `json:"routes"`
	}
	type Driving struct {
		Status int `json:"status"`
		Message string `json:"message"`
		Result Result `json:"result"`
	}

	ResData := Driving{}
	json.Unmarshal(content,&ResData)
	//对路线时间进行修正
	//记录总路段数
	step_num := len(ResData.Result.Routes[0].Steps)
	var congestion int  //拥堵的路段数
	//读取路段的拥堵情况
	for i := 0; i < step_num; i++ {
		if(ResData.Result.Routes[0].Steps[i].TrafficCondition[0].Status >= 3){
			congestion++
		}
	}
	//计算新的路线时间
	congestionl := float64(congestion)
	step_numl := float64(step_num)
	_duration := float64(ResData.Result.Routes[0].Duration)
	__duration := int(_duration*(congestionl/step_numl) + _duration*(1-congestionl/step_numl)*(1.4))
	//写入新的路线时间
	ResData.Result.Routes[0].Duration = __duration

	switch mode{
	case "1":
		fmt.Fprintf(w,"<h1>路线时间%d秒</h1><br>",ResData.Result.Routes[0].Duration)  //输出总耗时
	case "2":
		fmt.Fprintf(w,"<h1>路线时间%d秒</h1><br>",ResData.Result.Routes[0].Duration)
		for step,station := range ResData.Result.Routes[0].Steps{
			fmt.Fprintf(w,"%d.%s<br>",step + 1,station.Instruction)     //输出具体的路线
		}
	case "3":
		fmt.Fprintf(w,"<h1>路线时间%d秒</h1><br>",ResData.Result.Routes[0].Duration)
		for step,station := range ResData.Result.Routes[0].Steps{
			if step != 0{
				fmt.Fprintf(w,"->(%s,%s)",station.StartLocation.Lat,station.StartLocation.Lng)  //输出每一站的经纬度，实现形式化的路线输出
			}else{
				fmt.Fprintf(w,"(%s,%s)",station.StartLocation.Lat,station.StartLocation.Lng)
			}
		}
	}

	
}
func walking(w http.ResponseWriter,content []byte,mode string){

	type Origin struct {
		Lng float64 `json:"lng"`
		Lat float64 `json:"lat"`
	}
	type Destination struct {
		Lng float64 `json:"lng"`
		Lat float64 `json:"lat"`
	}
	type StartLocation struct {
		Lng string `json:"lng"`
		Lat string `json:"lat"`
	}
	type EndLocation struct {
		Lng string `json:"lng"`
		Lat string `json:"lat"`
	}
	type Steps struct {
		Distance int `json:"distance"`
		Duration int `json:"duration"`
		Direction int `json:"direction"`
		Instruction string `json:"instruction"`
		Path string `json:"path"`
		StartLocation StartLocation `json:"start_location"`
		EndLocation EndLocation `json:"end_location"`
	}
	type Routes struct {
		Distance int `json:"distance"`
		Duration int `json:"duration"`
		Steps []Steps `json:"steps"`
	}
	type Result struct {
		Origin Origin `json:"origin"`
		Destination Destination `json:"destination"`
		Routes []Routes `json:"routes"`
	}
	type Walking struct {
		Status int `json:"status"`
		Message string `json:"message"`
		Result Result `json:"result"`
	}
	ResData := Walking{}
	json.Unmarshal(content,&ResData)

	switch mode{
	case "1":
		fmt.Fprintf(w,"<h1>路线时间%d秒</h1><br>",ResData.Result.Routes[0].Duration)  //输出总耗时
	case "2":
		fmt.Fprintf(w,"<h1>路线时间%d秒</h1><br>",ResData.Result.Routes[0].Duration)
		for step,station := range ResData.Result.Routes[0].Steps{
			fmt.Fprintf(w,"%d.%s<br>",step + 1,station.Instruction)     //输出具体的路线
		}
	case "3":
		fmt.Fprintf(w,"<h1>路线时间%d秒</h1><br>",ResData.Result.Routes[0].Duration)
		for step,station := range ResData.Result.Routes[0].Steps{
			if step != 0{
				fmt.Fprintf(w,"->(%s,%s)",station.StartLocation.Lat,station.StartLocation.Lng)  //输出每一站的经纬度，实现形式化的路线输出
			}else{
				fmt.Fprintf(w,"(%s,%s)",station.StartLocation.Lat,station.StartLocation.Lng)
			}
		}
	}
	
}
func riding(w http.ResponseWriter,content []byte,mode string){
	type Origin struct {
		Lng float64 `json:"lng"`
		Lat float64 `json:"lat"`
	}
	type Destination struct {
		Lng float64 `json:"lng"`
		Lat float64 `json:"lat"`
	}
	type StartLocation struct {
		Lng string `json:"lng"`
		Lat string `json:"lat"`
	}
	type EndLocation struct {
		Lng string `json:"lng"`
		Lat string `json:"lat"`
	}
	type Steps struct {
		Distance int `json:"distance"`
		Duration int `json:"duration"`
		Direction int `json:"direction"`
		TurnType string `json:"turn_type"`
		Name string `json:"name"`
		Instruction string `json:"instruction"`
		RestrictionsInfo string `json:"restrictions_info"`
		Path string `json:"path"`
		StartLocation StartLocation `json:"start_location"`
		EndLocation EndLocation `json:"end_location"`
	}
	type Routes struct {
		Distance int `json:"distance"`
		Duration int `json:"duration"`
		Steps []Steps `json:"steps"`
	}
	type Result struct {
		Origin Origin `json:"origin"`
		Destination Destination `json:"destination"`
		Routes []Routes `json:"routes"`
	}
	type Riding struct {
		Status int `json:"status"`
		Message string `json:"message"`
		Result Result `json:"result"`
	}
	ResData := Riding{}
	json.Unmarshal(content,&ResData)
		//对路线时间进行修正

	switch mode{
	case "1":
		fmt.Fprintf(w,"<h1>路线时间%d秒</h1><br>",ResData.Result.Routes[0].Duration)  //输出总耗时
	case "2":
		fmt.Fprintf(w,"<h1>路线时间%d秒</h1><br>",ResData.Result.Routes[0].Duration)
		for step,station := range ResData.Result.Routes[0].Steps{
			fmt.Fprintf(w,"%d.%s<br>",step + 1,station.Instruction)     //输出具体的路线
		}
	case "3":
		fmt.Fprintf(w,"<h1>路线时间%d秒</h1><br>",ResData.Result.Routes[0].Duration)
		for step,station := range ResData.Result.Routes[0].Steps{
			if step != 0{
				fmt.Fprintf(w,"->(%s,%s)",station.StartLocation.Lat,station.StartLocation.Lng)  //输出每一站的经纬度，实现形式化的路线输出
			}else{
				fmt.Fprintf(w,"(%s,%s)",station.StartLocation.Lat,station.StartLocation.Lng)
			}
		}
	}
	
}
func transit(w http.ResponseWriter,content []byte,mode string){
	

	type Origin struct {
		Lng float64 `json:"lng"`
		Lat float64 `json:"lat"`
	}
	type Destination struct {
		Lng float64 `json:"lng"`
		Lat float64 `json:"lat"`
	}
	type LinePrice struct {
		LinePrice int `json:"line_price"`
		LineType int `json:"line_type"`
	}
	type Vehicle struct {
		DirectText string `json:"direct_text"`
		Name string `json:"name"`
		LineID string `json:"line_id"`
		StartName string `json:"start_name"`
		EndName string `json:"end_name"`
		StartTime string `json:"start_time"`
		EndTime string `json:"end_time"`
		StopNum int `json:"stop_num"`
		TotalPrice int `json:"total_price"`
		Type int `json:"type"`
		ZonePrice int `json:"zone_price"`
	}
	type StartLocation struct {
		Lng float64 `json:"lng"`
		Lat float64 `json:"lat"`
	}
	type EndLocation struct {
		Lng float64 `json:"lng"`
		Lat float64 `json:"lat"`
	}
	type Steps struct {
		Distance int `json:"distance"`
		Duration int `json:"duration"`
		Type int `json:"type"`
		Instruction string `json:"instruction"`
		Vehicle Vehicle `json:"vehicle"`
		Path string `json:"path"`
		StartLocation StartLocation `json:"start_location"`
		EndLocation EndLocation `json:"end_location"`
	}
	type Routes struct {
		Distance int `json:"distance"`
		Duration int `json:"duration"`
		Price int `json:"price"`
		LinePrice []LinePrice `json:"line_price"`
		Steps [][]Steps `json:"steps"`
		TrafficCondition int `json:"traffic_condition"`
	}
	type Detail struct {
		Desc string `json:"desc"`
		KmPrice float64 `json:"km_price"`
		StartPrice int `json:"start_price"`
		TotalPrice int `json:"total_price"`
	}
	type Taxi struct {
		Detail []Detail `json:"detail"`
		Distance int `json:"distance"`
		Duration int `json:"duration"`
		Remark string `json:"remark"`
	}
	type Result struct {
		Origin Origin `json:"origin"`
		Destination Destination `json:"destination"`
		Routes []Routes `json:"routes"`
		Taxi Taxi `json:"taxi"`
	}
	type Transit struct {
		Status int `json:"status"`
		Message string `json:"message"`
		Result Result `json:"result"`
	}
	ResData := Transit{}
	json.Unmarshal(content,&ResData)
	switch mode{
	case "1":
		fmt.Fprintf(w,"<h1>路线时间%d秒</h1><br>",ResData.Result.Routes[0].Duration)  //输出总耗时
	case "2":
		fmt.Fprintf(w,"<h1>路线时间%d秒</h1><br>",ResData.Result.Routes[0].Duration)
		for step,station := range ResData.Result.Routes[0].Steps{   //此处的steps是个二维数组
			fmt.Fprintf(w,"%d.%s<br>",step + 1,station[1].Instruction)
		}
	case "3":
		fmt.Fprintf(w,"<h1>路线时间%d秒</h1><br>",ResData.Result.Routes[0].Duration)
		for step,station := range ResData.Result.Routes[0].Steps{
			if step != 0{
				fmt.Fprintf(w,"->(%f,%f)",station[0].StartLocation.Lat,station[0].StartLocation.Lng)  //输出每一站的经纬度，实现形式化的路线输出
			}else{
				fmt.Fprintf(w,"(%f,%f)",station[0].StartLocation.Lat,station[0].StartLocation.Lng)
			}
		}
	}
}
