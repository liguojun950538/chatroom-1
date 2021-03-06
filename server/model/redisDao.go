package model

import (
	"chatroom/common/message"
	"encoding/json"
	"fmt"

	"github.com/go-redis/redis"
)

// data assace object

type RedisDAO struct {
	key string
}

// 根据用户id获取数据库里的用户信息
func (this *RedisDAO) getUserById(id uint32) (user *message.User, err error) {
	client := redisPool.Get()
	res, err := client.HGet(this.key, string(id)).Result()
	if err != nil { //需判断错误是否是因连接断开引起的还是未找到该用户
		//fmt.Println("client.HGet(this.key, string(userId)).Result() error:", err)
		if err == redis.Nil {
			err = ERROR_USER_NOTEXISTS
		} else {
			err = ERROR_SERVER
		}
		return
	}
	user = &message.User{}
	err = json.Unmarshal([]byte(res), user)
	if err != nil {
		//fmt.Println("user.Unserializer(res) error:", err)
		err = ERROR_USER_FORMAT
		return
	}
	return
}

func (this *RedisDAO) getUserCount() (count uint32, err error) {
	client := redisPool.Get()
	res, err := client.HLen(this.key).Result()
	if err != nil {
		err = ERROR_SERVER
		return
	}
	count = uint32(res)
	return
}

// 新的Id从100000开始
var baseId = uint32(100000)

func (this *RedisDAO) getNewUserId() (id uint32, err error) {
	count, err := this.getUserCount()
	if err != nil {
		return
	}
	id = count + baseId
	_, err = this.getUserById(id)
	if err != nil {
		return
	}
	return
}

func (this *RedisDAO) addUser(user *message.User) (err error) {
	client := redisPool.Get()
	data, err := json.Marshal(*user)
	err = client.HSet(this.key, string(user.UserId), data).Err()
	if err != nil {
		fmt.Println(" client.HSet(this.key, string(user.UserId), user).Result() error:", err)
		return
	}
	return
}

func (this *RedisDAO) Login(userId uint32, userPwd string) (user *message.User, err error) {
	user, err = this.getUserById(userId)
	if err != nil {
		fmt.Println("getUserById(userId) error:", err)
		return
	}
	fmt.Println("userPwd:", userPwd)
	fmt.Println("user:", user)
	if userPwd != user.UserPwd {
		user = nil
		err = ERROR_USER_PWD
		return
	}
	return
}

func (this *RedisDAO) Register(user *message.User) (err error) {
	//_, err = this.getUserById(user.UserId)
	user.UserId, err = this.getNewUserId()
	if err == ERROR_USER_NOTEXISTS { //数据库中无此账号id，可以注册
		err = this.addUser(user) //返回nil即注册成功
		return
	} else if err == nil { //用户已经存在
		err = ERROR_USER_EXISTS
	} else { //内部错误
		return
	}
	return
}
