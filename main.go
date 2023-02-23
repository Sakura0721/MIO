package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/Sakura0721/mio/config"
	"github.com/gotd/td/session"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/tg"
	"github.com/tidwall/gjson"
)

var (
	ErrQueryIDNotFound = fmt.Errorf("query_id not found in URL")
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func getQueryID(url string) (string, error) {
	r := regexp.MustCompile("query_id%3D(.*?)%26")
	rs := r.FindStringSubmatch(url)
	if len(rs) < 2 {
		return "", ErrQueryIDNotFound
	}
	return rs[1], nil
}

func getAuthDate(url string) (string, error) {
	r := regexp.MustCompile("auth_date%3D(.*?)%")
	rs := r.FindStringSubmatch(url)
	if len(rs) < 2 {
		return "", ErrQueryIDNotFound
	}
	return rs[1], nil
}

func getHash(url string) (string, error) {
	r := regexp.MustCompile("hash%3D(.*?)\\&")
	rs := r.FindStringSubmatch(url)
	if len(rs) < 2 {
		return "", ErrQueryIDNotFound
	}
	return rs[1], nil
}

type userData struct {
	Id           int64  `json:"id"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	Username     string `json:"username"`
	LanguageCode string `json:"language_code"`
}

type appData struct {
	QueryID  string   `json:"query_id"`
	User     userData `json:"user"`
	AuthDate string   `json:"auth_date"`
	Hash     string   `json:"hash"`
}

type loginRequest struct {
	Appdata appData `json:"appdata"`
}

func getCurrentUser() {

}

func getLoginRequest(queryID, authDate, hash string) []byte {
	r := loginRequest{
		Appdata: appData{
			QueryID: queryID,
			User: userData{
				Id:           config.C.User.ID,
				FirstName:    config.C.User.FirstName,
				LastName:     config.C.User.LastName,
				Username:     config.C.User.Username,
				LanguageCode: "en",
			},
			AuthDate: authDate,
			Hash:     hash,
		},
	}
	b, err := json.Marshal(r)
	check(err)
	fmt.Println(string(b))
	return b
}

type loginResponseData struct {
	AppSid string `json:"appSid"`
	Uid    string `json:"uid"`
}

type loginResponse struct {
	Data loginResponseData `json:"data"`
}

func login(queryID, authDate, hash string) {
	fmt.Println(queryID, authDate, hash)
	httpposturl := "https://service-bqh4il8q-1301162125.hk.apigw.tencentcs.com/webapp/TGWebAppLogin"
	fmt.Println("HTTP JSON POST URL:", httpposturl)

	request, error := http.NewRequest("POST", httpposturl, bytes.NewBuffer(getLoginRequest(queryID, authDate, hash)))
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")
	request.Header.Set("sec-ch-ua", "\"Chromium\";v=\"110\", \"Not A(Brand\";v=\"24\", \"Microsoft Edge\";v=\"110\", \"Microsoft Edge WebView2\";v=\"110\"")
	request.Header.Set("apiVer", "20221224")
	request.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/110.0.0.0 Safari/537.36 Edg/110.0.1587.50")

	client := &http.Client{}
	response, error := client.Do(request)
	if error != nil {
		panic(error)
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	check(err)
	r := loginResponse{}
	err = json.Unmarshal(body, &r)
	check(err)
	config.C.AppSid = r.Data.AppSid
	config.C.Uid = r.Data.Uid
}

func postData(url string, data []byte) []byte {
	url = fmt.Sprintf("https://service-bqh4il8q-1301162125.hk.apigw.tencentcs.com/%s", url)

	request, error := http.NewRequest("POST", url, bytes.NewBuffer(data))
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")
	request.Header.Set("appSid", config.C.AppSid)
	request.Header.Set("apiVer", "20221224")

	client := &http.Client{}
	response, error := client.Do(request)
	if error != nil {
		panic(error)
	}
	defer response.Body.Close()

	body, _ := ioutil.ReadAll(response.Body)
	return body
}

func addTime() {
	fmt.Println("正在为指定用户增加时间：")
	for _, username := range config.C.AddTime.Usernames {
		res := string(postData("webapp/FindUser", []byte(fmt.Sprintf("{\"type\":\"昵称\",\"opt\":{\"昵称\":\"%s\",\"条件\":{\"可以支配\":false,\"可以服从\":false}}}", username))))
		uids := gjson.Get(res, "data.玩家组.#.uid").Array()
		if len(uids) == 0 {
			fmt.Printf("user %s not found\n", username)
			continue
		}
		uid := uids[0].String()
		fmt.Println(username)
		fmt.Println(string(postData("shijiansuo/addTime", []byte(fmt.Sprintf("{\"uid\":\"%s\"}", uid)))))
	}
}

func replyAddTime() {
	fmt.Println("正在为给你增加过时间的用户增加时间")
	res := string(postData("webapp/getUserInfo", []byte{}))
	uids := gjson.Get(res, "data.TimeLock.时间变化记录.#(事件名==\"时间锁_增加时间\")#.uid").Array()
	ts := gjson.Get(res, "data.TimeLock.时间变化记录.#(事件名==\"时间锁_增加时间\")#.addtime").Array()
	now := time.Now()
	for i, uid := range uids {
		t := time.UnixMilli(ts[i].Int())
		if now.Sub(t) > time.Hour*8 {
			continue
		}
		fmt.Println(string(postData("shijiansuo/addTime", []byte(fmt.Sprintf("{\"uid\":\"%s\"}", uid.String())))))
	}
	return
}

func addTimeAll() {
	fmt.Println("正在给公开列表玩家增加时间")
	res := string(postData("shijiansuo/getLockedUsers", []byte{}))
	uids := gjson.Get(res, "data.公开.#.uid").Array()
	usernames := gjson.Get(res, "data.公开.#.玩家信息.昵称").Array()
	for i, uid := range uids {
		if _, ok := config.C.AddTimeAll.ExcludeUsernamesMap[usernames[i].String()]; ok {
			continue
		}
		fmt.Println(string(postData("shijiansuo/addTime", []byte(fmt.Sprintf("{\"uid\":\"%s\"}", uid.String())))))
	}
}

func giveHearts() {

	fmt.Println("正在给指定用户送心")
	for _, username := range config.C.GiveHearts.Usernames {
		res := string(postData("webapp/FindUser", []byte(fmt.Sprintf("{\"type\":\"昵称\",\"opt\":{\"昵称\":\"%s\",\"条件\":{\"可以支配\":false,\"可以服从\":false}}}", username))))
		uids := gjson.Get(res, "data.玩家组.#.uid").Array()
		if len(uids) == 0 {
			fmt.Printf("user %s not found\n", username)
			continue
		}
		uid := uids[0].String()
		fmt.Println(username)
		fmt.Println(string(postData("webapp/GiveHeart", []byte(fmt.Sprintf("{\"tuid\":\"%s\"}", uid)))))
	}
}

// replyHearts will give heart to all users who gave heart to you today.
func replyHearts() {
	fmt.Println("正在回赠心心")
	res := string(postData("webapp/getUserInfo", []byte{}))
	uids := gjson.Get(res, "data.近期事件组.#(事件名==\"收到爱心\")#.事件.uid").Array()
	ts := gjson.Get(res, "data.近期事件组.#(事件名==\"收到爱心\")#.addtime").Array()
	now := time.Now().Day()
	for i, uid := range uids {
		t := time.UnixMilli(ts[i].Int()).Day()
		if t != now {
			continue
		}
		fmt.Println(string(postData("webapp/GiveHeart", []byte(fmt.Sprintf("{\"tuid\":\"%s\"}", uid.String())))))
	}
	return
}

func main() {
	apiID := config.C.ApiId
	apiHash := config.C.ApiHash
	phone := config.C.Phone
	if apiID == 0 || apiHash == "" || phone == "" {
		log.Fatal("PHONE, APP_ID or APP_HASH is not set")
	}

	ctx := context.Background()
	client := telegram.NewClient(apiID, apiHash, telegram.Options{
		SessionStorage: &session.FileStorage{
			Path: "./gotd.session",
		},
	})
	codeAsk := func(ctx context.Context, sentCode *tg.AuthSentCode) (string, error) {
		fmt.Print("code:")
		code, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			return "", err
		}
		code = strings.ReplaceAll(code, "\n", "")
		return code, nil
	}

	check(client.Run(ctx, func(ctx context.Context) error {
		fmt.Println("正在登录")
		flow := auth.NewFlow(
			auth.CodeOnly(phone, auth.CodeAuthenticatorFunc(codeAsk)),
			auth.SendCodeOptions{},
		)
		check(client.Auth().IfNecessary(ctx, flow))
		id, err := client.RandInt64()
		check(err)
		api := client.API()

	fmt.Println("正在获取用户信息")

		currentUser, err := api.ContactsResolvePhone(ctx, config.C.Phone)
		check(err)
		self, ok := currentUser.Users[0].AsNotEmpty()
		if !ok {
			panic("get current user failed")
		}
		config.C.User = self

		peer, err := api.ContactsResolveUsername(ctx, "mioGameBot")
		check(err)
		mioPeer := tg.InputPeerUser{}

		for _, user := range peer.Users {
			u, ok := user.AsNotEmpty()
			if !ok {
				continue
			}
			if u.Username == "mioGameBot" {
				mioPeer.UserID = u.ID
				mioPeer.AccessHash = u.AccessHash
			}
		}

		fmt.Println("正在登录MIO")

		_, err = api.MessagesSendMessage(context.Background(), &tg.MessagesSendMessageRequest{
			Flags:                  0,
			NoWebpage:              true,
			Silent:                 false,
			Background:             false,
			ClearDraft:             true,
			Noforwards:             false,
			UpdateStickersetsOrder: false,
			Peer:                   &mioPeer,
			TopMsgID:               0,
			Message:                "/start",
			RandomID:               id,
			ReplyMarkup:            nil,
			Entities:               []tg.MessageEntityClass{},
			ScheduleDate:           0,
			SendAs:                 nil,
		})
		check(err)
		startParam := fmt.Sprintf("t=%d", time.Now().UnixMilli())
		msgID := 0
		for {
			time.Sleep(time.Second * 10)
			msgs, err := api.MessagesGetHistory(ctx, &tg.MessagesGetHistoryRequest{
				Peer:  &mioPeer,
				Limit: 1,
			})
			check(err)
			if strings.Contains(msgs.String(), "/start") {
				continue
			}
			r := regexp.MustCompile("\\?(t=.*?)\\}")
			rs := r.FindStringSubmatch(msgs.String())
			startParam = rs[1]
			msgSlice := msgs.(*tg.MessagesMessagesSlice)
			msgID = msgSlice.Messages[0].(*tg.Message).ID
			break
		}

		upd, err := api.MessagesRequestWebView(ctx, &tg.MessagesRequestWebViewRequest{
			Flags:        0,
			FromBotMenu:  false,
			Silent:       false,
			Peer:         &mioPeer,
			Bot:          &tg.InputUser{UserID: mioPeer.UserID, AccessHash: mioPeer.AccessHash},
			URL:          fmt.Sprintf("https://webapp.mio-game.com/WEBAPP/APP/WEB/"),
			StartParam:   startParam,
			ThemeParams:  tg.DataJSON{},
			Platform:     "Windows",
			ReplyToMsgID: msgID,
			TopMsgID:     msgID,
			SendAs:       nil,
		})
		check(err)
		queryID, err := getQueryID(upd.URL)
		check(err)

		authDate, err := getAuthDate(upd.URL)
		check(err)

		hash, err := getHash(upd.URL)
		check(err)

		login(queryID, authDate, hash)

		if config.C.GiveHearts.Enabled {
			giveHearts()
		}

		if config.C.ReplyHearts.Enabled {
			replyHearts()
		}

		if config.C.AddTime.Enabled {
			addTime()
		}

		if config.C.AddTimeAll.Enabled {
			addTimeAll()
		}

		if config.C.ReplyAddTime.Enabled {
			replyAddTime()
		}
		// for {
		// 	fmt.Println(api.MessagesProlongWebView(ctx, &tg.MessagesProlongWebViewRequest{
		// 		Flags:        0,
		// 		Silent:       false,
		// 		Peer:         &mioPeer,
		// 		Bot:          &tg.InputUser{UserID: mioPeer.UserID, AccessHash: mioPeer.AccessHash},
		// 		QueryID:      upd.QueryID,
		// 		ReplyToMsgID: 0,
		// 		TopMsgID:     0,
		// 		SendAs:       nil,
		// 	}))
		// 	time.Sleep(time.Second * 50)
		// }
		return nil
	}))

}
