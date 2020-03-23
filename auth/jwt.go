package auth

import (
	jwt "github.com/dgrijalva/jwt-go"
	"time"
)

// 这是加密秘钥
var jwtSecret = []byte("2020wqservice")

type Claims struct {
	Username string `json:"username"`
	Telephone string `json:"telephone"`
	jwt.StandardClaims
}

// 利用时间戳生成token
func GenerateToken(username, telephone string) (string, error) {
	nowTime := time.Now()
	expireTime := nowTime.Add(3 * time.Hour)

	claims := Claims{
		username,
		telephone,
		jwt.StandardClaims{
			Subject:   "",
			ExpiresAt: expireTime.Unix(),
			Issuer:    "weqiservice",
		},
	}
	tokenClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err := tokenClaims.SignedString(jwtSecret)
	return token, err
}

func ParseToken(token string) (*Claims, error) {
	tokenClaims, err := jwt.ParseWithClaims(token, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if tokenClaims != nil {
		if claims, ok := tokenClaims.Claims.(*Claims); ok && tokenClaims.Valid {
			return claims, nil
		}
	}
	return nil, err
}
