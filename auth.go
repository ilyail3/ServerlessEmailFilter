package main

import (
	"google.golang.org/appengine/datastore"
	"encoding/hex"
	"net/http"
	"github.com/dgrijalva/jwt-go"
	"fmt"
	"errors"
	"context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/user"
	"encoding/json"
)

type AuthUser interface {
	Email() string
	Id() string
	IsAdmin() bool
}

type jwtUser struct {
	email string
	id string
	admin bool
}

func(u *jwtUser)Email() string{
	return u.email
}

func(u *jwtUser)Id() string{
	return u.id
}

func(u *jwtUser)IsAdmin() bool{
	return u.admin
}

type secretStruct struct {
	Data string
}

type loginUrlStruct struct {
	Url string
}

func secret(ctx context.Context)([]byte, error) {
	key := datastore.NewKey(ctx, "secret", "secret", 0, nil)
	data := secretStruct{}

	err := datastore.Get(ctx, key, &data)

	if err != nil {
		return nil, err
	}

	bytes, err := hex.DecodeString(data.Data)

	if err != nil {
		return nil, err
	}

	return bytes, err
}

func TokenAuth(ctx context.Context, r *http.Request)(AuthUser, error){
	if r.Header.Get("X-JWT") == "" {
		return nil,nil
	}

	token, err := jwt.Parse(r.Header.Get("X-JWT"), func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		secret, err := secret(ctx)

		if err != nil {
			return nil, err
		}

		// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
		return secret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return &jwtUser{
			email: claims["email"].(string),
			id: claims["id"].(string),
			admin: claims["admin"].(bool),
		}, nil
	} else {
		return nil, errors.New("token doesn't seem to be valid")
	}
}

func LoginHandle(w http.ResponseWriter, r *http.Request){
	ctx := appengine.NewContext(r)
	u := user.Current(ctx)
	if u == nil {

		url, _ := user.LoginURL(ctx, r.URL.Path)

		w.Header().Set("Content-type", "application/json; charset=utf-8")
		encoder := json.NewEncoder(w)

		encoder.Encode(map[string]string{
			"url": url,
		})

		return
	}

	jwtSecret,err := secret(ctx)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id": u.ID,
		"email": u.Email,
		"admin": u.Admin,
	})

	tokenString, err := token.SignedString(jwtSecret)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-type", "text/plain; charset=utf-8")
	fmt.Fprint(w,tokenString)
}