package keycloak

import (
	"fmt"
	"github.com/getevo/evo"
	"github.com/getevo/evo/lib/log"
	"gopkg.in/square/go-jose.v2/jwt"
)

type User struct {
	Exp               int64  `json:"exp"`
	IAT               int64  `json:"iat"`
	JTI               string `json:"jti"`
	ISS               string `json:"iss"`
	Sub               string `json:"sub"`
	Typ               string `json:"typ"`
	Azp               string `json:"azp"`
	SessionState      string `json:"session_state"`
	Acr               string `json:"acr"`
	Scope             string `json:"scope"`
	EmailVerified     bool   `json:"email_verified"`
	Name              string `json:"name"`
	PreferredUsername string `json:"preferred_username"`
	GivenName         string `json:"given_name"`
	FamilyName        string `json:"family_name"`
	Email             string `json:"email"`
}

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
	u.Name = p.Name
	u.Email = p.Email
	u.Username = p.PreferredUsername
	u.Name = p.Name + " " + p.FamilyName
	return nil
}
func (p User) SyncPermissions(app string, perms evo.Permissions) {
	log.Info("implement me")
}

// SetGroup set user group
func (p User) FromRequest(request *evo.Request) {
	request.User = &evo.User{Anonymous: true}
	accessToken := request.Get("access_token")
	if accessToken == "" {
		accessToken = request.Cookies("access_token")
	}
	if accessToken != "" {
		token, err := jwt.ParseSigned(accessToken)
		if err != nil {
			log.Error(err)
			request.WriteResponse(false, fmt.Errorf("unauthorized"), 401)
		}
		var claims User
		err = token.Claims(Certificates.Keys[0], &claims)
		if err != nil {
			log.Error(err)
			request.WriteResponse(false, fmt.Errorf("unauthorized"), 401)
		}

		request.User.Name = claims.Name + " " + claims.FamilyName
		request.User.Anonymous = false
		request.User.Email = claims.Email
		request.User.Username = claims.PreferredUsername
	}
}
