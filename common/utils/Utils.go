package utils

import (
	"crypto/sha256"
	"fmt"
	"github.com/gioco-play/easy-i18n/i18n"
	"github.com/shopspring/decimal"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/text/language"
	"math/rand"
	"reflect"
	"strings"
	"time"
)

type RandomType int8
type UppLowType int8

const (
	ALL    RandomType = 0
	NUMBER RandomType = 1
	STRING RandomType = 2
)

const (
	MIX   UppLowType = 0
	UPPER UppLowType = 1
	LOWER UppLowType = 2
)

// PasswordHash 密码加密
func PasswordHash(plainpwd string) string {
	//谷歌的加密包
	hash, err := bcrypt.GenerateFromPassword([]byte(plainpwd), bcrypt.DefaultCost) //加密处理
	if err != nil {
		fmt.Println(err)
	}
	encodePWD := string(hash) // 保存在数据库的密码，虽然每次生成都不同，只需保存一份即可
	return encodePWD
}

// CheckPassword 密码校验
func CheckPassword(plainpwd, cryptedpwd string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(cryptedpwd), []byte(plainpwd)) //验证（对比）
	return err == nil
}

func PasswordHash2(plainpwd string) string {
	sum := sha256.Sum256([]byte(plainpwd))
	return fmt.Sprintf("%x", sum)
}

func CheckPassword2(plainpwd string, cryptedpwd string) bool {
	sum := sha256.Sum256([]byte(plainpwd))
	fmt.Sprintf("aaaa", sum)
	//encodePWD := string(sum[:])
	if fmt.Sprintf("%x", sum) == cryptedpwd {
		return true
	} else {
		return false
	}
}

// ParseTime 時間隔式處理
func ParseTime(t string) string {
	timeString, err := time.Parse(time.RFC3339, t)
	if err != nil {
	}
	str := strings.Split(timeString.String(), " +")
	res := str[0]
	return res
}

// ParseTime
func ParseTimeAddOneSecond(t string) string {
	timeString, err := time.Parse("2006-01-02 15:04:05", t)
	if err != nil {
	}
	str := strings.Split(timeString.Add(time.Second*1).String(), " +")
	res := str[0]
	return res
}

// ParseIntTime int時間隔式處理
func ParseIntTime(t int64) string {
	return time.Unix(t, 0).UTC().Format("2006-01-02 15:04:05")
}

// Contain 判斷obj是否在target中，target支援的型別array,slice,map
func Contain(obj interface{}, target interface{}) bool {
	targetValue := reflect.ValueOf(target)
	switch reflect.TypeOf(target).Kind() {
	case reflect.Slice, reflect.Array:
		for i := 0; i < targetValue.Len(); i++ {
			if targetValue.Index(i).Interface() == obj {
				return true
			}
		}
	case reflect.Map:
		if targetValue.MapIndex(reflect.ValueOf(obj)).IsValid() {
			return true
		}
	}

	return false
}

//FloatMul 浮點數乘法 (precision=4)
func FloatMul(s float64, p float64, precisions ...int32) float64 {

	f1 := decimal.NewFromFloat(s)
	f2 := decimal.NewFromFloat(p)

	var precision int32
	if len(precisions) > 0 {
		precision = precisions[0]
	} else {
		precision = 3
	}

	res, _ := f1.Mul(f2).Truncate(precision).Float64()

	return res
}

//FloatDiv 浮點數除法 (precision=4)
func FloatDiv(s float64, p float64, precisions ...int32) float64 {

	f1 := decimal.NewFromFloat(s)
	f2 := decimal.NewFromFloat(p)

	var precision int32
	if len(precisions) > 0 {
		precision = precisions[0]
	} else {
		precision = 3
	}
	res, _ := f1.Div(f2).Truncate(precision).Float64()

	return res
}

//FloatSub 浮點數減法 (precision=4)
func FloatSub(s float64, p float64, precisions ...int32) float64 {

	f1 := decimal.NewFromFloat(s)
	f2 := decimal.NewFromFloat(p)

	var precision int32
	if len(precisions) > 0 {
		precision = precisions[0]
	} else {
		precision = 3
	}
	res, _ := f1.Sub(f2).Truncate(precision).Float64()

	return res
}

//FloatAdd 浮點數加法 (precision=4)
func FloatAdd(s float64, p float64, precisions ...int32) float64 {

	f1 := decimal.NewFromFloat(s)
	f2 := decimal.NewFromFloat(p)

	var precision int32
	if len(precisions) > 0 {
		precision = precisions[0]
	} else {
		precision = 3
	}
	res, _ := f1.Add(f2).Truncate(precision).Float64()

	return res
}

func SetI18n(languageX string) {
	if len(languageX) > 0 {
		i18n.SetLang(language.Make(languageX))
	} else {
		i18n.SetLang(language.English)
	}
}

//GetRandomString 生成随机字符串
func GetRandomString(length int, randomType RandomType, uppLowType UppLowType) string {
	var str string

	switch randomType {
	case NUMBER:
		str = "0123456789"
	case STRING:
		str = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	default:
		str = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	}

	bytes := []byte(str)
	result := []byte{}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < length; i++ {
		result = append(result, bytes[r.Intn(len(bytes))])
	}

	switch uppLowType {
	case UPPER:
		str = strings.ToUpper(str)
	case LOWER:
		str = strings.ToLower(str)
	}

	return string(result)
}
