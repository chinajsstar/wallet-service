package main

import (
	"flag"
	"fmt"
	"os"
	"bastionpay_api/utils"
	"encoding/base64"
	"bastionpay_api/api"
	"encoding/json"
)

var (
	clientPrivKey = `-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEA2ShdNBAr6APhtqEQMRgJ33nI6K5UfT49+YM3nCYcbslJhHgL
uUETCNufx7fY6IXExai/zuVXXUFm1AoBYkCwhuvDeYlecdtZiGmgTR20GvFf3Hag
Ywca7ZR5oSwi4fTCHq2+U7wf5qw0JZRUwY87PPUaO6Wv77Obx7SIV947w4tjjWN7
+ygkjDrUUDNwt7lGetXmkPDOs5a+mfFf3TyyzVZkEwJZZgur7ndvwRwTVpBZ+x7n
ZBggIyTrkwFSItMxy3btjbHMuenqkmwGoNpZmjmpaPqTrw0JIQdzbxph4EuUIS9J
A9x+eEwzzToJBRSzcJaNefomP6yQDuSFyXQjDQIDAQABAoIBADT60Qpjq6KWV9oL
n3yqxbXc63RBG+HWbp5SMh4JekRZHXORKiMPSkqN8oRySRwpjqE+k1UxxMe+rgyr
SD0lVSwFlxIuvnj+r/BE/NPznD46h02tL2IZmKs/3xDASN5hrX54mweozQulFa/Z
aXgzrpsnnTfSK4NKiYYGeIEfeesxwlZYyAJAoryw8BGL11r3P7CjVytpfYvNO8Iw
vm4DvP7bcNfWY+HL0TCicIlAldGMnULB69EnC/M5gRXE2t0sntLusi8m1sblLxT+
0HpAmUO7CzJXd+31xX0BcMq09kynYCpmOY5t2VZffBKHPWNKAMVGs3tGOwwuKloI
CxPeVM0CgYEA+12t1zX2zLHJ0G7XtanXh2u8sUj7SDlvl/jPBAtEdqPocOiJrq3a
9uwGFs1SccAVll3UNoNJPMfBfyRF9m2FLJJqPkb+XzHxxi+sjxgv9vrjA19ZOcwK
PAjzV4rGNm2CFbDqthPVYNH2fR7KGFlAaQSxv+qRpeQs+0iQil6WQ3MCgYEA3Sk9
PDOspOSdGOk4nr/2bAsw+LclHFXkO193DAZzeLjRLMTgtfz+j5GOGbrRih0Pp00O
xL8jofrNfdUG87t2FHj1CNgczCT5T7FQypDcquQNgu9xDr2AfxNz1434J4QNWfgV
PwNxUhD0eLwOwTCqlMGPpPDOymK/ePn5JmBbX38CgYAfQV29VdtzRmQWw8GUuCKx
lwbmcHG2YFXs6qYrFY+UxPrBF1kPXFTOwThcm3k15bGJee9KO/beVORNf+AnLSUT
Dr2hMsisY5RxlTn6rQJBzp5tq+x3UFSxdFd1ui69U86KDe6RZ1Pv7wucMTjl4Csv
I8NKXstvejjRogs++VXr2wKBgQCnSEEkHyXwF2foZsfH9AUNVi21groUQ7d9XUkr
lFmVL54wUb5yiSl3KK6mMZO69x5W+kc4dCccpxr1mcNHajz3YUJebTDjDPhy2fj1
LztqG1NI/Zmzf40JEYqmuaDBKvX6Xlabshvt6hswk3ZJlGkCWaIwLWEM3kJb3CAj
cfDLOwKBgGmtRzSMy3dDuSbYnpjRK9awHN5li4Yd8ym2OetfXXCyIF/uCNWKA99C
mPooCfDxhscpYeRiBFz9VNZbRX46yUm4hZMCJffVkEGYmy47Gtjo+4MzHA1znWCz
y4Cq5pR0/A5zwhN8FerUbcQug/G8qorif2CwkFjbklQG0t+ziGM5
-----END RSA PRIVATE KEY-----`

	serverPrivKey = `-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEAy7FVkwTaZkCHZi6xwvN+dlMhzzHUwTDZwpA4ubNQOMgKDazA
WwY5roXa8JY+3kuaDjnWq/5m5e57Fa55bavFFQI8N+Nn+Xrg/fcYMp3oQyQAr2a1
EfD88ipNm360vAwKHd5grYC+wVadqhH5QKZNj7GQRhe1pN9B7N2o7tCRIaFNac0p
OK7tythwQf2yo4foVvqbxhYrmArXS3GG7H6BxfDor4fC84Eqmrb4XHIKZYQQwLr6
4c9Ex10gumklDWkqCJPD7QX0dW7bCJA4NWnTSijJSRePNnAhJ6h94172WZp7gPk7
OSSW/P1ZByKyBrwuFHS98+L83aZQ3P2tW/+nzQIDAQABAoIBAHRE8GTwUz4kvWT9
aGiwltIx182uUxRHrzVZH5gYm8Udomd14YLTxd8x9ux3xsEbBH/0Ue5xVAkRnN6e
bh/E+cVpNjhsrTACDSXKgtx5uFeC2IVGqjrohWox7YTByabBdJDiG+tN3xT7PRoU
EmPtyb4pDAKGjB/ldHshd8mB8iXwKCzTGbnI1a82Pk9ia3Dol918JiezHl+deK3u
hgDTX2DKZH9EeW/g+5kTUr2lkkniiRevZhBsvRfIEkDbw0Z5svHjJ6HnSbwQXugC
UO6SMEAjDT0guhQw94rfTQT2jlrvc1ECwIRAq+uk5i+ytV7GPTmGvw2LiZ+/wlRy
4HJg6eECgYEA0AbgiMujHw88vcMbR/iaScpEnAYR52RAAgLKa1OUmunfzaOT3LDo
Zjrmun7hvW5Q4K2QogV47kr2eRPy0jpBGLTvpH57F6D3IR2/EhbS7q2V8ew8s6fk
BblUwjTV8cPs40l55FXqEE+jY7JhTQsk9t0V6knKHUK10eLmDTz/95kCgYEA+qqV
zyOQ6bAsb4zqPV6sYktKBX/b7oF97SoAF2ElMMXQZomDZ6g8rISqPgJn3Tt9fECu
+nH8v77gG+r9mbZxO6s7/DIFZNZslCkKlbvqWtOxsCuJIkTYTU5Ld5pkO67wHcfX
lG3bvRz8qIZiMn1jxJmUE8NtSnmDpziFpT61QlUCgYB3LZlxhYi8gJRB+wckIm7y
G2lXIbscH7jz7fldp5KZdad8Ply1sLxT5SbObWaSiiLXtVgJGq1/h37ROvaALlOg
/ffU+4k1rkgmts4CZQUPLG+dG8RjAKqIBPdkE7UGP0L1q+CjSf1Avv9SXJ2V7+6L
Z3noscGTJebYie9WYj4a8QKBgQCIffbRMgaomRu5KjwqarDnXSPTVsoFV1GgoKwo
DufXZP+TtfFtsjhHoH2rAlhYKRqtW/NrwEHmjyMtUIoC90s1OIqTSxGQ1QmOq8Jt
wkNcbcVlrm8fz+kQPz/swo8tyJZLQRRsaF2s3mndwj8aSxjWZuIw2MtcMEq19Zsg
XcMJeQKBgBtxiYUg7zzz/18Vx69swO45d0Z5K6aepPrFHyYtsUcoHLF/jTUXKqkj
mlPISmhyZ7KDmTgzf9R5rXSI+wJ3+y2DtKUYLOAOKsOugPQcEpa+GTL+9/6/abGQ
2VwzQgzZ/P93eK6bDfpXud/yCdeXRqSSCk4rtobdGgCWjnvbHGgy
-----END RSA PRIVATE KEY-----`
)

func main() {
	var privkeyString *string = flag.String("key", clientPrivKey,
		"Use -key <private key to decrypto data>")

	var chiperString *string = flag.String("chiper", "",
		"Use -chiperString <base64 string of crypted data>")

	flag.Parse()

	fmt.Printf(`
------------------------------
private key : %s,
------------------------------
chiper base64:
%s
------------------------------
`, *privkeyString, *chiperString)

	chiper, err := base64.StdEncoding.DecodeString(*chiperString)
	if err != nil {
		fmt.Printf("error base64: %s\n", err.Error())
		os.Exit(-1)
		return
	}

	if false {
		ackData := api.UserResponseData{}
		if err:=json.Unmarshal(chiper, &ackData); err!=nil {
			fmt.Printf("unmarshal faild, message:%s\n", err.Error())
			os.Exit(-1)
			return
		}
		chiper = []byte(ackData.Value.Message)
	}

	plain, err := utils.RsaDecrypt([]byte(chiper),
		[]byte(*privkeyString), utils.RsaDecodeLimit2048)

	if err!=nil {
		fmt.Printf("decrypto faild, message:%s\n", err.Error())
		os.Exit(-1)
		return
	}

	fmt.Printf("plain text is : %s\n", string(plain))
}
