package keycloak

import (
	"fmt"
	"github.com/getevo/evo"
	"github.com/getevo/evo/lib/data"
	"github.com/getevo/evo/lib/log"
	"gopkg.in/square/go-jose.v2/jwt"
)

type User struct{}

func (p User) Save(u *evo.User) error {
	log.Info("implement me")
	return nil
}
func (p User) HasPerm(u *evo.User, v string) bool {
	log.Info("implement me")
	return true
}
func (p User) HasRole(u *evo.User, v interface{}) bool {
	log.Info("implement me")
	return true
}
func (p User) Image(u *evo.User) string {
	return "files/profile/profile-" + fmt.Sprint(u.ID) + ".jpg"
}
func (p User) SetPassword(u *evo.User, password string) error {
	log.Info("implement me")
	return nil
}
func (p User) SetGroup(u *evo.User, group interface{}) error {
	log.Info("implement me")
	return nil

}
func (p User) AfterFind(u *evo.User) error {

	return nil
}
func (p User) SyncPermissions(app string, perms evo.Permissions) {
	log.Info("implement me")
}

// SetGroup set user group
func (p User) FromRequest(request *evo.Request) {
	request.User = &evo.User{Anonymous: true}
	accessToken := request.Get("Authorization")
	if accessToken == "" {
		accessToken = request.Cookies("Authorization")
	}
	if accessToken != "" {
		if accessToken[0:6] == "Bearer" {
			accessToken = accessToken[7:]
		}

		token, err := jwt.ParseSigned(accessToken)
		if err != nil {
			//log.Error(err)
			//request.WriteResponse(false, fmt.Errorf("unauthorized"), 401)
			return
		}
		var claims data.Map
		err = token.Claims(Certificates.Keys[0], &claims)

		if err != nil {
			//log.Error(err)
			//request.WriteResponse(false, fmt.Errorf("unauthorized"), 401)
			return
		}
		claims.ToStruct(request.User)
		if claims.Get("user_id") != nil {
			request.User.Anonymous = false
			request.User.ID = uint(claims.Get("user_id").(float64))
		} else {
			request.User.Anonymous = true
		}
		request.User.Params = claims

	}
}
