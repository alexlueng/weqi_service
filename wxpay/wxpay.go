package wxpay

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
	"weqi_service/api"
	"weqi_service/models"
	"weqi_service/serializer"
)

type OrderRespData struct {
	AppId     string `json:"appId"`
	Timestamp string `json:"timeStamp"`
	NonceStr  string `json:"nonceStr"`
	Package   string `json:"package"`
	SignType  string `json:"signType"`
	PaySign   string `json:"paySign"`
	TotalFee  int    `json:"totalFee"`
	OrderNo   string `json:"order_no"`
}

type UserIDService struct {
	ModuleID int64 `json:"module_id"`
	UserID   int64 `json:"id"` //json字段名要改成user_id
}

func GetOpenIDURL(c *gin.Context) {

	//url := "https://open.weixin.qq.com/connect/oauth2/authorize?appid=" + os.Getenv("APPID") + "&redirect_uri=" + os.Getenv("OPENIDREDIRECTURL") + "&response_type=code&scope=snsapi_base&state=123"
	url := fmt.Sprintf(os.Getenv("GETOPENIDURL"), os.Getenv("APPID"), os.Getenv("OPENIDREDIRECTURL"))
	fmt.Println("url: ", url)

	c.JSON(http.StatusOK, serializer.Response{
		Code: 200,
		Msg:  "查询用户openID",
		Data: url,
	})
}

type WXPayUserOpenID struct {
	UserID int64 `json:"user_id"`
	InstanceID int64 `json:"instance_id"`
	ModuleName string `json:"module_name"`
	Price float64 `json:"price"`
	OpenID string `json:"openid"`
}

func GetPrepayID(c *gin.Context) {
	var openID WXPayUserOpenID
	if err := c.ShouldBindJSON(&openID); err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "无法得到code",
		})
		return
	}

	if openID.UserID == 0 || openID.InstanceID == 0 {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "缺乏必要参数，下单不成功",
		})
		return
	}


	//GetTestSignKey()
	// 生成一个订单
	order := models.Order{}
	order.OrderID = api.GetLatestID("order")
	order.OrderNO = api.GetTempOrderSN()
	order.InstanceID = openID.InstanceID
	order.UserID = openID.UserID
	order.Amount = openID.Price
	order.ModuleName = openID.ModuleName

	api.SmartPrint(order)

	var result interface{}

	ip := api.GetIpAddress(c)
	if IsMobile(c.Request.UserAgent()) {
		result = JSAPIPay(openID.OpenID, ip, order)
	} else {
		result = NativeGetShopCartForm(order.InstanceID, ip, order)
	}


	// 收到预支付订单后再将此订单写入数据库中
	collection := models.Client.Collection("orders")
	insertReseult, err := collection.InsertOne(context.TODO(), order)
	if err != nil {
		fmt.Println("Can't insert order: ", err)
		return
	}
	fmt.Println("insert order result: ", insertReseult.InsertedID)

	c.JSON(http.StatusOK, serializer.Response{
		Code: 200,
		Msg:  "get prepay id",
		Data: &result,
	})
}

func Callback(c *gin.Context) {
	code := c.Query("code")

	fmt.Println("get open id: ", code)

	if code == "" {
		fmt.Println("can't get code")
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "无法得到code",
		})
		return
	}

	openid := GetOpenID(code)
	fmt.Println("open id code: ", code)
	c.Redirect(http.StatusMovedPermanently, os.Getenv("REDIRECTURL") + openid)

	//otp-O1Vt3-fFPRSleTx2u9pUlnG0
}

// 每个User都有一个唯一的openID 需要把他存到数据库中
func GetOpenID(code string) string {
	//req_openid := "https://api.weixin.qq.com/sns/oauth2/access_token?appid=" + os.Getenv("APPID") + "&secret=" + os.Getenv("APPSECRET") + "&code=" + code + "&grant_type=authorization_code"
	req_openid := fmt.Sprintf(os.Getenv("REQOPENIDURL"), os.Getenv("APPID"), os.Getenv("APPSECRET"), code)
	//new request
	req, err := http.NewRequest(http.MethodGet, req_openid, nil)
	if err != nil {
		log.Println(err)
		return ""
	}

	//http client
	client := &http.Client{}
	log.Printf("Go %s URL : %s \n", http.MethodGet, req.URL.String())

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Can't set get request: ", err)
		return ""
	}

	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("ioutil.ReadAll %v", err)
		return ""
	}
	fmt.Println(string(respBytes))

	openID := WXPayUserOpenID{}
	err = json.Unmarshal(respBytes, &openID)
	if err != nil {
		fmt.Println("Can't get openid: ", err)
		return ""
	}
	api.SmartPrint(openID)
	return openID.OpenID
	//	GetShopCartForm(code)
}

/*
<xml>
   <appid>wx2421b1c4370ec43b</appid>
   <attach>支付测试</attach>
   <body>JSAPI支付测试</body>
   <mch_id>10000100</mch_id>
   <detail><![CDATA[{ "goods_detail":[ { "goods_id":"iphone6s_16G", "wxpay_goods_id":"1001", "goods_name":"iPhone6s 16G", "quantity":1, "price":528800, "goods_category":"123456", "body":"苹果手机" }, { "goods_id":"iphone6s_32G", "wxpay_goods_id":"1002", "goods_name":"iPhone6s 32G", "quantity":1, "price":608800, "goods_category":"123789", "body":"苹果手机" } ] }]]></detail>
   <nonce_str>1add1a30ac87aa2db72f57a2375d8fec</nonce_str>
   <notify_url>http://wxpay.wxutil.com/pub_v2/pay/notify.v2.php</notify_url>
   <openid>oUpF8uMuAJO_M2pxb1Q9zNjWeS6o</openid>
   <out_trade_no>1415659990</out_trade_no>
   <spbill_create_ip>14.23.150.211</spbill_create_ip>
   <total_fee>1</total_fee>
   <trade_type>JSAPI</trade_type>
   <sign>0CB01533B8C1EF103065174F50BCA001</sign>
</xml>
*/

//  1.2   微信支付计算签名的函数
func wxpayCalcSign(mReq map[string]interface{}, key string) (sign string) {
	fmt.Println("微信支付签名计算, API KEY:", key)
	//STEP 1, 对key进行升序排序.
	sorted_keys := make([]string, 0)
	for k, _ := range mReq {
		sorted_keys = append(sorted_keys, k)
	}

	sort.Strings(sorted_keys)

	//STEP2, 对key=value的键值对用&连接起来，略过空值
	var signStrings string
	for _, k := range sorted_keys {
		fmt.Printf("k=%v, v=%v\n", k, mReq[k])
		value := fmt.Sprintf("%v", mReq[k])
		if value != "" {
			signStrings = signStrings + k + "=" + value + "&"
		}
	}

	//STEP3, 在键值对的最后加上key=API_KEY
	if key != "" {
		signStrings = signStrings + "key=" + key
	}
	//STEP4, 进行MD5签名并且将所有字符转为大写.
	md5Ctx := md5.New()
	md5Ctx.Write([]byte(signStrings))
	cipherStr := md5Ctx.Sum(nil)
	upperSign := strings.ToUpper(hex.EncodeToString(cipherStr))

	fmt.Println("Get wxpay sign: ", upperSign)

	return upperSign
}

	//
func NativeGetShopCartForm(moduleID int64, ip string, order models.Order) map[string]interface{} {
	var WxUnifiedReq models.NativeUnifyOrderReq
	WxUnifiedReq.Appid = os.Getenv("APPID")
	WxUnifiedReq.ProductID = moduleID
	WxUnifiedReq.Body = order.ModuleName + "模块续费"
	WxUnifiedReq.Mch_id = os.Getenv("MCHID")// 商户号
	cardNumber := api.GenRandomDigitCode(15)
	WxUnifiedReq.Nonce_str = cardNumber //随机字符串
	WxUnifiedReq.Notify_url = os.Getenv("WXPAYCALLBACKURL")
	WxUnifiedReq.Trade_type = "NATIVE"
	WxUnifiedReq.Spbill_create_ip = ip
	price, err := strconv.Atoi(os.Getenv("PRICEOFJXC"))
	if err != nil {
		fmt.Println("Can' get the price: ", err)
	}
	WxUnifiedReq.Total_fee = price //价格，【单位：分】
	WxUnifiedReq.Out_trade_no = order.OrderNO
	//WxUnifiedReq.Openid = openID

	var m map[string]interface{}
	m = make(map[string]interface{}, 0)
	m["appid"] = WxUnifiedReq.Appid
	m["body"] = WxUnifiedReq.Body
	m["mch_id"] = WxUnifiedReq.Mch_id
	m["notify_url"] = WxUnifiedReq.Notify_url
	m["trade_type"] = WxUnifiedReq.Trade_type
	m["spbill_create_ip"] = WxUnifiedReq.Spbill_create_ip
	m["total_fee"] = WxUnifiedReq.Total_fee
	m["out_trade_no"] = WxUnifiedReq.Out_trade_no
	//m["openid"] = WxUnifiedReq.Openid
	m["nonce_str"] = WxUnifiedReq.Nonce_str
	m["product_id"] = WxUnifiedReq.ProductID
	//微信支付下单签名
	//WxUnifiedReq.Sign = wxpayCalcSign(m, "SXLejYrYAPlq5Q4XzKwMJOIMs76e7ofk") //这个key 微信商户平台(pay.weixin.qq.com)-->账户设置-->API安全
	WxUnifiedReq.Sign = wxpayCalcSign(m, os.Getenv("MCHKEY"))
	//fmt.Println("-----------------微信支付下单签名-----------------", yourReq.Sign)
	bytes_req, err := xml.Marshal(WxUnifiedReq)
	if err != nil {
		fmt.Println("以xml形式编码发送错误, 原因:", err)
		return nil
	}
	str_req := string(bytes_req)
	//wxpay的unifiedorder接口需要http body中xmldoc的根节点是<xml></xml>这种，所以这里需要replace一下
	str_req = strings.Replace(str_req, "UnifyOrderReq", "xml", -1)
	bytes_req = []byte(str_req)
	//发送unified order请求.
	req, err := http.NewRequest("POST", os.Getenv("WXORDERURL"), bytes.NewReader(bytes_req))
	if err != nil {
		fmt.Println("New Http Request发生错误，原因:", err)
		return nil
	}
	req.Header.Set("Accept", "application/xml")
	//这里的http header的设置是必须设置的.
	req.Header.Set("Content-Type", "application/xml;charset=utf-8")
	client := http.Client{}
	resp, _err := client.Do(req)
	if _err != nil {
		fmt.Println("请求微信支付统一下单接口发送错误, 原因:", _err)
		return nil
	}
	//到这里统一下单接口就已经执行完成了
	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("解析返回body错误", err)
		return nil
	}

	xmlResp := models.NativeUnifyOrderResp{}
	_err = xml.Unmarshal(respBytes, &xmlResp)
	fmt.Println(string(respBytes))

	//处理return code.
	if xmlResp.Return_code == "FAIL" {
		fmt.Println("微信支付统一下单不成功，原因:", xmlResp.Return_msg, " str_req-->", str_req)
		return nil
	}
	//这里已经得到微信支付的prepay id，需要返给客户端，由客户端继续完成支付流程
	fmt.Println("微信支付统一下单成功，结果:", xmlResp.Return_msg, " str_req-->", str_req)
	fmt.Println("微信支付统一下单成功，预支付单号:", xmlResp.Prepay_id)
	fmt.Println("微信支付统一下单成功，二维码链接:", xmlResp.Code_url)

	//向数据库存入 WxBill WxUnifiedReq.Out_trade_no user的wx_id orderId 时间 total_fee
	timeStamp := time.Now().Unix()
	//paySign := getPaySign(&WxUnifiedReq, &xmlResp, timeStamp)
	timeStamps := strconv.FormatInt(timeStamp, 10)

	//fmt.Println("------------------------>返回前端的参数paySign<-------------：", paySign)

	var res = map[string]interface{} {
		"prepay_id":  xmlResp.Prepay_id,
		"code_url":   xmlResp.Code_url,
		"appId":      WxUnifiedReq.Appid,
		"timeStamp":  timeStamps,
		"OutTradeNo": WxUnifiedReq.Out_trade_no,
		"nonceStr":   WxUnifiedReq.Nonce_str,
		"price": WxUnifiedReq.Total_fee,
		//"signature":getTicketSigNature(&Req),
		//"paySign":paySign,
	}

	return res
}

func JSAPIPay(openID string, ip string, order models.Order) *OrderRespData {
	var WxUnifiedReq models.UnifyOrderReq
	WxUnifiedReq.Appid = os.Getenv("APPID")
	WxUnifiedReq.Body = "JSAPI支付测试"
	WxUnifiedReq.Mch_id = os.Getenv("MCHID") // 商户号
	cardNumber := api.GenRandomDigitCode(15)
	WxUnifiedReq.Nonce_str = cardNumber //随机字符串
	WxUnifiedReq.Notify_url = os.Getenv("WXPAYCALLBACKURL")
	WxUnifiedReq.Trade_type = "JSAPI"
	WxUnifiedReq.Spbill_create_ip = ip
	fmt.Println("device ip: ", ip)

	WxUnifiedReq.Total_fee = 1 //价格，【单位：分】
	WxUnifiedReq.Out_trade_no = order.OrderNO
	WxUnifiedReq.Openid = openID

	var m map[string]interface{}
	m = make(map[string]interface{}, 0)
	m["appid"] = WxUnifiedReq.Appid
	m["body"] = WxUnifiedReq.Body
	m["mch_id"] = WxUnifiedReq.Mch_id
	m["notify_url"] = WxUnifiedReq.Notify_url
	m["trade_type"] = WxUnifiedReq.Trade_type
	m["spbill_create_ip"] = WxUnifiedReq.Spbill_create_ip
	m["total_fee"] = WxUnifiedReq.Total_fee
	m["out_trade_no"] = WxUnifiedReq.Out_trade_no
	m["openid"] = WxUnifiedReq.Openid
	m["nonce_str"] = WxUnifiedReq.Nonce_str
	//微信支付下单签名
	//WxUnifiedReq.Sign = wxpayCalcSign(m, "SXLejYrYAPlq5Q4XzKwMJOIMs76e7ofk") //这个key 微信商户平台(pay.weixin.qq.com)-->账户设置-->API安全
	WxUnifiedReq.Sign = wxpayCalcSign(m, os.Getenv("MCHKEY"))
	//WxUnifiedReq.Sign = "f4cede9126f4a5afad1fe2ffc510a95c"
	bytes_req, err := xml.Marshal(WxUnifiedReq)
	if err != nil {
		fmt.Println("以xml形式编码发送错误, 原因:", err)
		return nil
	}
	str_req := string(bytes_req)
	//wxpay的unifiedorder接口需要http body中xmldoc的根节点是<xml></xml>
	str_req = strings.Replace(str_req, "UnifyOrderReq", "xml", -1)
	bytes_req = []byte(str_req)
	//发送unified order请求.
	req, err := http.NewRequest("POST", os.Getenv("WXORDERURL"), bytes.NewReader(bytes_req))
	if err != nil {
		fmt.Println("New Http Request发生错误，原因:", err)
		return nil
	}
	req.Header.Set("Accept", "application/xml")
	//这里的http header的设置是必须设置的.
	req.Header.Set("Content-Type", "application/xml;charset=utf-8")
	client := http.Client{}
	resp, _err := client.Do(req)
	if _err != nil {
		fmt.Println("请求微信支付统一下单接口发送错误, 原因:", _err)
		return nil
	}
	//到这里统一下单接口就已经执行完成了
	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("解析返回body错误", err)
		return nil
	}
	xmlResp := models.UnifyOrderResp{}
	_err = xml.Unmarshal(respBytes, &xmlResp)
	fmt.Println(string(respBytes))
	//处理return code.
	if xmlResp.Return_code == "FAIL" {
		fmt.Println("微信支付统一下单不成功，原因:", xmlResp.Return_msg, " str_req-->", str_req)
		return nil
	}
	//这里已经得到微信支付的prepay id，需要返给客户端，由客户端继续完成支付流程
	fmt.Println("微信支付统一下单成功，结果:", xmlResp.Return_msg, " str_req-->", str_req)
	fmt.Println("微信支付统一下单成功，预支付单号:", xmlResp.Prepay_id)
	//向数据库存入 WxBill WxUnifiedReq.Out_trade_no user的wx_id orderId 时间 total_fee
	timeStamp := time.Now().Unix()
	paySign := getPaySign(&WxUnifiedReq, &xmlResp, timeStamp)
	timeStamps := strconv.FormatInt(timeStamp, 10)

	fmt.Println("------------------------>返回前端的参数paySign<-------------：", paySign)

	resData := OrderRespData{
		AppId:     WxUnifiedReq.Appid,
		Timestamp: timeStamps,
		NonceStr:  WxUnifiedReq.Nonce_str,
		Package:   "prepay_id=" + xmlResp.Prepay_id,
		SignType:  "MD5",
		PaySign:   paySign,
		TotalFee:  WxUnifiedReq.Total_fee,
		OrderNo: order.OrderNO,
	}
	api.SmartPrint(resData)
	return &resData
}

//获取支付签名 跟下单签名不同的地方在于 最后一个字符串连接没有&【MD5】  1
func getPaySign(yourReq *models.UnifyOrderReq, xmlResp *models.UnifyOrderResp, timeStamp int64) string {
	p := make(map[string]interface{}, 0)
	p["appId"] = yourReq.Appid
	p["timeStamp"] = timeStamp
	p["nonceStr"] = yourReq.Nonce_str
	p["package"] = "prepay_id=" + xmlResp.Prepay_id
	p["signType"] = "MD5"
	//return wxpaySign(p, "SXLejYrYAPlq5Q4XzKwMJOIMs76e7ofk")
	return wxpaySign(p, os.Getenv("MCHKEY"))
}

//计算支付签名 跟下单签名不同的地方在于 最后一个字符串连接没有&  2
func wxpaySign(mReq map[string]interface{}, key string) string {
	//STEP 1, 对key进行升序排序.
	sorted_keys := make([]string, 0)
	for k, _ := range mReq {
		sorted_keys = append(sorted_keys, k)
	}
	sort.Strings(sorted_keys)
	//STEP2, 对key=value的键值对用&连接起来，略过空值
	var signStrings string
	for i, k := range sorted_keys {
		value := fmt.Sprintf("%v", mReq[k])
		if value != "" {
			if i != (len(sorted_keys) - 1) {
				signStrings = signStrings + k + "=" + value + "&"
			} else {
				signStrings = signStrings + k + "=" + value //最后一个不加此符号
			}
		}
	}
	//STEP3, 在键值对的最后加上key=API_KEY
	if key != "" {
		signStrings = signStrings + "&key=" + key
	}
	fmt.Println("=====wxpaySign 键值对加key==============", signStrings)
	//STEP4, 进行MD5签名并且将所有字符转为大写.
	md5Ctx := md5.New()
	md5Ctx.Write([]byte(signStrings))
	cipherStr := md5Ctx.Sum(nil)
	upperSign := strings.ToUpper(hex.EncodeToString(cipherStr))
	fmt.Println("=====计算支付签名 【与返回前端的参数paySign作比较】==============", upperSign)
	return upperSign
}

// 微信异步通知
// 异步通知的数据结构
type WXPayNotifyReq struct {
	Return_code    string `xml:"return_code"`
	Return_msg     string `xml:"return_msg"`
	Appid          string `xml:"appid"`
	Mch_id         string `xml:"mch_id"`
	Nonce          string `xml:"nonce_str"`
	Sign           string `xml:"sign"`
	Result_code    string `xml:"result_code"`
	Openid         string `xml:"openid"`
	Is_subscribe   string `xml:"is_subscribe"`
	Trade_type     string `xml:"trade_type"`
	Bank_type      string `xml:"bank_type"`
	Total_fee      int    `xml:"total_fee"`
	Fee_type       string `xml:"fee_type"`
	Cash_fee       int    `xml:"cash_fee"`
	Cash_fee_Type  string `xml:"cash_fee_type"`
	Transaction_id string `xml:"transaction_id"`
	Out_trade_no   string `xml:"out_trade_no"`
	Attach         string `xml:"attach"`
	Time_end       string `xml:"time_end"`
}

type WXPayNotifyResp struct {
	Return_code string `xml:"return_code"`
	Return_msg  string `xml:"return_msg"`
}

// 微信支付回调函数
func WxpayCallback(c *gin.Context) {
	//配置请求头
	c.Header("Access-Control-Allow-Origin", "*")
	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		fmt.Println("读取http body失败，原因!", err)
		return
	}
	defer c.Request.Body.Close()

	fmt.Println("微信支付异步通知，HTTP Body:", string(body))
	var mr WXPayNotifyReq
	err = xml.Unmarshal(body, &mr)
	if err != nil {
		fmt.Println("解析HTTP Body格式到xml失败，原因!", err)
		return
	}

	var reqMap map[string]interface{}
	reqMap = make(map[string]interface{}, 0)

	reqMap["return_code"] = mr.Return_code
	reqMap["return_msg"] = mr.Return_msg
	reqMap["appid"] = mr.Appid
	reqMap["mch_id"] = mr.Mch_id
	reqMap["nonce_str"] = mr.Nonce
	reqMap["result_code"] = mr.Result_code
	reqMap["openid"] = mr.Openid
	reqMap["is_subscribe"] = mr.Is_subscribe
	reqMap["trade_type"] = mr.Trade_type
	reqMap["bank_type"] = mr.Bank_type
	reqMap["total_fee"] = mr.Total_fee
	reqMap["fee_type"] = mr.Fee_type
	reqMap["cash_fee"] = mr.Cash_fee
	reqMap["cash_fee_type"] = mr.Cash_fee_Type
	reqMap["transaction_id"] = mr.Transaction_id
	reqMap["out_trade_no"] = mr.Out_trade_no
	reqMap["attach"] = mr.Attach
	reqMap["time_end"] = mr.Time_end

	var resp WXPayNotifyResp
	//进行签名校验
	if wxpayVerifySign(reqMap, mr.Sign) {
		//微信支付成功
		resp.Return_code = "SUCCESS"
		resp.Return_msg = "OK"

		// 更新后台数据库
		// 根据订单号更新订单状态，如果交易成功，则同步到付费历史表,并更新用户该模块的过期时间
		// 1 更新订单状态
		orderNo := reqMap["out_trade_no"]
		fmt.Println("orderNo: ", orderNo)
		collection := models.Client.Collection("orders")
		var order models.Order
		err := collection.FindOne(context.TODO(), bson.D{{"order_no", orderNo}}).Decode(&order)
		if err != nil {
			fmt.Println("Can't find orders: ", err)
			return
		}
		updateResult, err := collection.UpdateOne(context.TODO(), bson.D{{"order_no", orderNo}}, bson.M{"$set": bson.M{"status": 1}})
		if err != nil {
			fmt.Println("Can't update order: ", err)
			return
		}
		fmt.Println("update result: ", updateResult.UpsertedID)
		api.SmartPrint(order)
		// 2 同步付费历史表
		var paidHistory models.PaidHistory
		paidHistory.ID = api.GetLatestID("paid_history")
		paidHistory.UserID = order.UserID
		paidHistory.InstanceID = order.InstanceID
		paidHistory.Amount = order.Amount
		paidHistory.PaidAt = time.Now().Unix()
		collection = models.Client.Collection("paid_history")

		insertResult, err := collection.InsertOne(context.TODO(), paidHistory)
		if err != nil {
			fmt.Println("Can't insert user paid history: ", err)
			return
		}
		fmt.Println("insert result: ", insertResult.InsertedID)

		// 3 更新模块实例过期时间
		collection = models.Client.Collection("module_instance")
		var moduleIns models.ModuleInstance
		err = collection.FindOne(context.TODO(), bson.D{{"id", order.InstanceID}}).Decode(&moduleIns)
		if err != nil {
			fmt.Println("Can't find module instance: ", err)
			return
		}
		api.SmartPrint(moduleIns)
		updateResult, err = collection.UpdateOne(context.TODO(), bson.D{{"id", moduleIns.InstanceID}}, bson.M{"$set": bson.M{
			"last_paid_at" : time.Now().Unix(),
			"expire_at" : moduleIns.ExpireAt + 365 * 86400, // 过期时间延长一年，暂时写死
		}})
		if err != nil {
			fmt.Println("Can't update module instance: ", err)
			return
		}
		fmt.Println("update result: ", updateResult.UpsertedID)

	} else {
		resp.Return_code = "FAIL"
		resp.Return_msg = "failed to verify sign, please retry!"
	}

	//结果返回，微信要求如果成功需要返回return_code "SUCCESS"
	byte_str, _err := xml.Marshal(resp)
	strResp := strings.Replace(string(byte_str), "WXPayNotifyResp", "xml", -1)
	if _err != nil {
		fmt.Println("xml编码失败，原因：", _err)
		return
	}

	c.String(http.StatusOK, strResp)

}

//  3.4  微信支付签名验证函数 【不含sign】
func wxpayVerifySign(needVerifyM map[string]interface{}, sign string) bool {
	//signCalc := wxpayCalcSign(needVerifyM, "SXLejYrYAPlq5Q4XzKwMJOIMs76e7ofk")
	signCalc := wxpayCalcSign(needVerifyM, os.Getenv("MCHKEY"))
	if sign == signCalc {
		fmt.Println("签名校验通过!")
		return true
	}
	fmt.Println("签名校验失败!")
	return false
}

func IsMobile(userAgent string) bool {
	if len(userAgent) == 0 {
		return false
	}

	isMobile := false
	mobileKeywords := []string{"Mobile", "Android", "Silk/", "Kindle",
		"BlackBerry", "Opera Mini", "Opera Mobi"}

	for i := 0; i < len(mobileKeywords); i++ {
		if strings.Contains(userAgent, mobileKeywords[i]) {
			isMobile = true
			break
		}
	}

	return isMobile
}

type TradeNoService struct {
	OutTradeNo string `json:"trade_no"`
}

func CheckOrderStatus(c *gin.Context) {
	var tradeNo TradeNoService
	if err := c.ShouldBindJSON(&tradeNo); err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "无法得到trade no",
		})
		return
	}
	api.SmartPrint(tradeNo)
	var order models.Order
	collection := models.Client.Collection("orders")
	err := collection.FindOne(context.TODO(), bson.D{{"order_no", tradeNo.OutTradeNo}}).Decode(&order)
	if err != nil {
		fmt.Println("Can't find order: ", err)
		return
	}
	c.JSON(http.StatusOK, serializer.Response{
		Code: -1,
		Msg:  "trade no",
		Data: order,
	})

}


func UserPaidHistory(c *gin.Context) {
	var uPaidHistorySrv UserIDService
	if err := c.ShouldBindJSON(&uPaidHistorySrv); err != nil {
		c.JSON(http.StatusOK, serializer.Response{
			Code: -1,
			Msg:  "无法得到trade no",
		})
		return
	}
	api.SmartPrint(uPaidHistorySrv)
	collection := models.Client.Collection("paid_history")
	var histories []models.PaidHistory
	filter := bson.M{}
	filter["user_id"] = uPaidHistorySrv.UserID
	cur, err := collection.Find(context.TODO(), filter)
	if err != nil {
		fmt.Println("Can't find history: ", err)
		return
	}
	for cur.Next(context.TODO()) {
		var res models.PaidHistory
		if err := cur.Decode(&res); err != nil {
			fmt.Println("Can't decode paid hisory: ", err)
			return
		}
		histories = append(histories, res)
	}
	c.JSON(http.StatusOK, serializer.Response{
		Code: 200,
		Msg:  "User paid history",
		Data: histories,
	})
}

