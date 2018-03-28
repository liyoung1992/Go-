package main

import (
	"log"
	"net/http"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/gomodule/redigo/redis"
	"encoding/json"
	"time"
	"strconv"
)

//websocket连接和uid做映射
var conn_uid = make(map[int](*websocket.Conn)) 
// redis 连接对象
var redis_conn  redis.Conn
// var redis_ip string= "192.168.1.188:6379"
// var redis_auth_pwd string = "1234"
var redis_ip string= "127.0.0.1:6379"

var redis_auth_pwd string = "123456"

// Configure the upgrader
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Define our message object
type Message struct {
	Uid int `json:"uid"`
	Message  string `json:"message"`
	Sendtime int `json:"send_time"`
	// Readtime string `json:"read_time"`
	// Deletetime string `json:"delete_time"`
	Receiveruid int `json:"receiver_uid"`
}

func main() {
	// Create a simple file server
	fs := http.FileServer(http.Dir("../admin"))
	http.Handle("/", fs)

	// redis连接
	var r_err error
	redis_conn,r_err = redis.Dial("tcp",string(redis_ip))   
	if r_err != nil {
		log.Fatal("ListenAndServe: ", r_err)
	}
	redis_conn.Do("AUTH",string(redis_auth_pwd))
	
	// 客户连接处理
	http.HandleFunc("/ws", handleClientConn)

	//客服登陆处理
	http.HandleFunc("/custom_service",handleServerConn)
	// 获取所有的消息列表
	http.HandleFunc("/user_msg_list",get_user_list)
	// 获取消息详情
	http.HandleFunc("/msg_info",get_msg_by_uid)

	// Start the server on localhost port 8000 and log any errors
	log.Println("http server started on :8000")
	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
func sendMessageByUid(w http.ResponseWriter, r *http.Request) {

}
// 处理客服登陆
func handleServerConn(w http.ResponseWriter, r *http.Request) {
	log.Printf("handleServerConn.....")
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("客服连接失败.....")
		log.Fatal(err)
	}
	// 关闭websocket
	defer ws.Close()
    conn_uid[1] = ws

   // 监听客户发送的消息
	for {
		var msg Message
	    //读取客服的消息
		err := ws.ReadJSON(&msg)
		// msg.Receiveruid = -1
		msg.Sendtime = int(time.Now().UnixNano()/1000000)
		uid := msg.Uid
		//uid,_ := strconv.Atoi(msg.Uid)
		// log.Printf("server msg:", uid)
		log.Printf("send client msg:", uid)
		if err != nil {
			log.Printf("error: %v", err)
			delete(conn_uid, uid)
			break
		}
		// msg.Receiveruid = 1001
		transport_msg(uid,msg.Receiveruid,msg)
	}
}
// 处理客户登陆
func handleClientConn(w http.ResponseWriter, r *http.Request) {
	log.Printf("client connect........")
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	// 关闭websocket
	defer ws.Close()

	var msg Message
	// 读取用户的id，并把用户的信息写入redis
	err_uid := ws.ReadJSON(&msg)
	
	msg.Receiveruid = 1
	// msg.Receiveruid = -1
	msg.Sendtime = int(time.Now().UnixNano()/1000000)
	// redis_set(msg)
	if err_uid != nil {
		log.Printf("error: %v", err_uid)
		return
	}
	// broadcast <- uid_msg
	uid := msg.Uid
	//uid,_ := strconv.Atoi(msg.Uid)
	log.Printf("client msg uid:",uid)
	conn_uid[uid]=  ws
	log.Printf("send msg uid:",uid)
	transport_msg(uid,msg.Receiveruid,msg)
   // 监听客户发送的消息
	for {
		var msg Message
		//读取客户消息
		log.Printf("读取客户消息.....")
		err := ws.ReadJSON(&msg)
		if err != nil {
			log.Printf("error: %v", err)
			delete(conn_uid, uid)
			break
		}
		//用户发来的消息
		msg.Receiveruid = 1
		msg.Sendtime = int(time.Now().UnixNano()/1000000)
		//uid,_ := strconv.Atoi(msg.Uid)
		uid := msg.Uid
		log.Printf("send server msg:", uid)
		if err != nil {
			log.Printf("error: %v", err)
			delete(conn_uid, uid)
			break
		}
		// }
	
		transport_msg(uid,msg.Receiveruid,msg)
		redis_set(msg)
	}
}
// 获取用户列表
func get_user_list(w http.ResponseWriter, r *http.Request) {
	log.Printf("get_user_list start")
	type UserMessage struct {
		Uid int `json:"uid"`
		Len int `json:"len"`
	}
	keys ,_ :=  redis.Values(redis_conn.Do("keys","*"))
	user_list := make([]UserMessage,0,10) 
	for _,v := range keys {
			var msg UserMessage
			msg.Uid,_ = strconv.Atoi(string(v.([]byte)))
			msg.Len,_ = redis.Int(redis_conn.Do("llen",v))
			if msg.Uid > 10 {
				user_list = append(user_list,msg)
			}	
	}
	b,_ := json.Marshal(user_list)	
	log.Printf("get_user_list end")
	fmt.Fprint(w,string(b))
}

//获取消息详情
func get_msg_by_uid(w http.ResponseWriter, r *http.Request){
	// fmt.Fprintln(w,r.Form["uid"][0])
	r.ParseForm() 
	log.Println(r.Form)
	// fmt.Fprint(w,r.Form.Get("uid"))
 	values ,_ := redis.Values(redis_conn.Do("lrange",r.Form.Get("uid"),0,-1))
	msg_list := make([]Message,0,10)
	for _,v1 := range values {
		var msg Message
		fmt.Println(string(v1.([]byte)))
		json.Unmarshal(v1.([]byte), &msg)		
		msg_list = append(msg_list,msg)
	}
	redis_conn.Do("del",r.Form.Get("uid"))
	b,_ := json.Marshal(msg_list)	
	fmt.Fprint(w,string(b))
}
//聊天信息写入redis
func redis_set(m Message) {
	b,err := json.Marshal(m)
	if err != nil {
		log.Printf("error: %v", err)
	}
	var uid = m.Uid;
	redis_conn.Do("lpush",uid,string(b))
}

//消息传输（包括客户到客服）
func transport_msg(from_id int,to_id int,m Message){
	log.Printf("send msg to id:", to_id)
	_,ok := conn_uid[to_id]
	if ok {
		err := conn_uid[to_id].WriteJSON(m)
		if err != nil {
			// 用户不在线，写入redis
			redis_set(m)
			log.Printf("error: %v", err)
			conn_uid[to_id].Close()
			delete(conn_uid, to_id)
		}
	}else {
		// 用户不在线，写入redis
		redis_set(m)
		log.Printf("用户不在线")
	}
}

