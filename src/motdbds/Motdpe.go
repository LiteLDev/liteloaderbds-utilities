package main

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type ServerInfo struct {
	Status      string
	EditionStr  string
	Motd        string
	Protocol    string
	VersionName string
	PlayerCount int
	MaxPlayers  int
	UniqueId    string
	Motd2       string
	GamemodeStr string
	GamemodeInt int
	PortV4      int
	PortV6      int
	PingMs      int
}
type Config struct {
	HttpAddr string
}

func motdpe(host, port string) (info ServerInfo, err error) {
	addr, err := net.ResolveUDPAddr("udp", host+":"+port)
	if err != nil {
		return
	}
	time1 := time.Now().UnixNano()
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return
	}
	_ = conn.SetDeadline(time.Now().Add(5 * time.Second))

	defer conn.Close()
	bin, _ := hex.DecodeString("01010000000000000000ffff00fefefefefdfdfdfd12345678")
	_, err = conn.Write(bin)
	if err != nil {
		return
	}
	_ = conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	data := make([]byte, 4096)
	length, err := conn.Read(data)
	if err != nil {
		err = errors.New("Failed To Read Response Or Server Offline")
		return
	}
	time2 := time.Now().UnixNano()
	ret := strings.Split(string(data[35:length]), ";")
	info.EditionStr = ret[0]
	info.Motd = ret[1]
	info.Protocol = ret[2]
	info.VersionName = ret[3]
	info.PlayerCount, _ = strconv.Atoi(ret[4])
	info.MaxPlayers, _ = strconv.Atoi(ret[5])
	info.UniqueId = ret[6]
	info.Motd2 = ret[7]
	info.GamemodeStr = ret[8]
	info.GamemodeInt, _ = strconv.Atoi(ret[9])
	info.PortV4, _ = strconv.Atoi(ret[10])
	info.PortV6, _ = strconv.Atoi(ret[11])
	info.PingMs = int((time2 - time1) / 2000000)
	info.Status = "Online"
	return
}
func LoadConfig() Config {
	f, err := os.OpenFile("motdpe.json", os.O_RDONLY, 0600)
	defer f.Close()
	if err != nil {
		log.Printf("[Error] Error when reading config > %s", err.Error())
	}
	config_bytes, _ := ioutil.ReadAll(f)
	var config_json Config
	config_json.HttpAddr = "0.0.0.0:8080"
	err = json.Unmarshal(config_bytes, &config_json)
	if err != nil {
		log.Printf("[Error] Something wrong with motdpe.json > %s", err.Error())
	}
	return config_json
}
func main() {
	quitSignChan := make(chan os.Signal)
	cfg := LoadConfig()
	http.HandleFunc("/motdpe", MotdpeHandler)
	http.HandleFunc("/", IndexHandler)
	log.Printf("HttpServer started listen: http://%s", cfg.HttpAddr)
	err := http.ListenAndServe(cfg.HttpAddr, nil)
	if errors.Is(err, http.ErrServerClosed) {
		return
	}
	if err != nil {
		quitSignChan <- syscall.SIGKILL
		log.Fatalf("HttpServer start fail: %v", err)
	}
}
func IndexHandler(respw http.ResponseWriter, req *http.Request) {
	//fmt.Println(req.URL.Path)
	if req.URL.Path == "/" {
		respw.Write([]byte("<!DOCTYPE html>\n<html>\n<head>\n<meta http-equiv=\"Content-Type\" content=\"text/html; charset=utf-8\" />\n<title>MotdPe Service</title>\n</head>\n<body>\n<big> MotdPe Services For MCPE</big><br>\n<input type=\"text\" id='IpAddr' placeholder=\"IPAddress\" style=\"width:150px;height:20px\">\n<input type=\"text\" id='Port' placeholder=\"Port\" style=\"width:50px;height:20px\">\n<br>\n<button id=\"btn\" onclick=\"serch();\" type=\"submit\" style=\"width:215px;height:30px\"/><img alt=\"GetServerInfo\" /></button>\n<script>\nfunction serch() {\n\tvar IP = document.getElementById('IpAddr').value;\n\tvar Port = document.getElementById('Port').value;\n\tif (document.getElementById('Port').value == '') {\n\t\tPort = '19132'\n\t}\n\twindow.open('motdpe?ip=' + IP + '&port=' + Port);\n}\n</script>\n</body>\n</html>\n"))
	}
	return
}

func MotdpeHandler(respw http.ResponseWriter, req *http.Request) {
	query, _ := url.ParseQuery(req.URL.RawQuery)
	var ip string
	port := "19132"
	if len(query["ip"]) == 0 || query["ip"][0] == "" {
		fmt.Fprintf(respw, "{\"Status\":\"Error: Please provide ip\"}")
		return
	}
	ip = query["ip"][0]
	if len(query["port"]) != 0 {
		port = query["port"][0]
	}
	info, err := motdpe(ip, port)
	if err != nil {
		fmt.Fprintf(respw, "{\"Status\":\"Error: %s\"}", err.Error())
		return
	}
	rets, _ := json.Marshal(info)
	respw.Write(rets)
	return
}
