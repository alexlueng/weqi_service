package models

//  1.1   首先定义一个UnifyOrderReq用于填入我们要传入的参数。
type UnifyOrderReq struct {
	Appid            string `xml:"appid"`            //公众账号ID
	//Attach           string `xml:"attach"`
	Detail           string `xml:"detail"`
	Body             string `xml:"body"`             //商品描述
	Mch_id           string `xml:"mch_id"`           //商户号
	Nonce_str        string `xml:"nonce_str"`        //随机字符串
	Notify_url       string `xml:"notify_url"`       //微信支付结果通知地址
	Trade_type       string `xml:"trade_type"`       //交易类型
	Spbill_create_ip string `xml:"spbill_create_ip"` //支付提交用户端ip
	Total_fee        int    `xml:"total_fee"`        //总金额【分】
	Out_trade_no     string `xml:"out_trade_no"`     //商户订单号
	Sign             string `xml:"sign"`             //签名

	Openid           string `xml:"openid"`           //购买商品的用户wxid
}

type NativeUnifyOrderReq struct {
	Appid            string `xml:"appid"`            //公众账号ID
	//Attach           string `xml:"attach"`
	ProductID        int64 `xml:"product_id"`
	Detail           string `xml:"detail"`
	Body             string `xml:"body"`             //商品描述
	Mch_id           string `xml:"mch_id"`           //商户号
	Nonce_str        string `xml:"nonce_str"`        //随机字符串
	Notify_url       string `xml:"notify_url"`       //微信支付结果通知地址
	Trade_type       string `xml:"trade_type"`       //交易类型
	Spbill_create_ip string `xml:"spbill_create_ip"` //支付提交用户端ip
	Total_fee        int    `xml:"total_fee"`        //总金额【分】
	Out_trade_no     string `xml:"out_trade_no"`     //商户订单号
	Sign             string `xml:"sign"`             //签名

	Openid           string `xml:"openid"`           //购买商品的用户wxid
}



type UnifyOrderResp struct {
	Return_code string `xml:"return_code"`
	Return_msg  string `xml:"return_msg"`
	Appid       string `xml:"appid"`
	Mch_id      string `xml:"mch_id"`
	Nonce_str   string `xml:"nonce_str"`
	Sign        string `xml:"sign"`
	Result_code string `xml:"result_code"`
	Prepay_id   string `xml:"prepay_id"`
	Trade_type  string `xml:"trade_type"`
	Mweb_url    string `xml:"mweb_url"`
}


type NativeUnifyOrderResp struct {
	Return_code string `xml:"return_code"`
	Return_msg  string `xml:"return_msg"`
	Appid       string `xml:"appid"`
	Mch_id      string `xml:"mch_id"`
	Nonce_str   string `xml:"nonce_str"`
	Sign        string `xml:"sign"`
	Result_code string `xml:"result_code"`
	Prepay_id   string `xml:"prepay_id"`
	Trade_type  string `xml:"trade_type"`
	Code_url    string `xml:"code_url"`
}