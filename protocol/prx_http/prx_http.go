package prx_http

func MSG_ProxyAuthRequired() string {
	return "" +
		"HTTP/1.1 407 Proxy Authentication Required\r\n" +
		"Proxy-Authenticate: Basic realm=\"GoProxy\"\r\n\r\n"
}

func MSG_BadRequest() string {
	return "" +
		"HTTP/1.1 400 Bad Request\r\n\r\n"
}

func MSG_BadGateway() string {
	return "" +
		"HTTP/1.1 502 Bad Gateway\r\n\r\n"
}

func MSG_ConnectionEtablished() string {
	return "" +
		"HTTP/1.1 200 Connection Established\r\n\r\n"
}
