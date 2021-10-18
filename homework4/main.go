package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/golang-jwt/jwt"
)

type idType string

const userID idType = "ID"

type MessgBody struct {
	Message string `json:"message"`
}

type MessgFrom struct {
	ID   idType
	Text string
}

type MessgList struct {
	ID       idType
	Messages []string
}

var jwtKey = []byte("some_key")

var Users = map[idType]string{}

type Credentials struct {
	Username idType `json:"username"`
	Password string `json:"password"`
}

type Claims struct {
	Username idType `json:"username"`
	jwt.StandardClaims
}

type FromTo struct {
	Sender    idType
	Recipient idType
}

var PrivateMessages = map[FromTo][]string{}

var GroupMessages []MessgFrom

func main() {
	root := chi.NewRouter()
	root.Use(middleware.Logger)
	root.Post("/users/register", Register)
	root.Post("/users/signin", SignIn)

	r := chi.NewRouter()
	r.Use(Auth)
	r.Get("/messages", GetMessages)
	r.Post("/messages", PostMessages)
	r.Get("/users/me/messages", GetPrivateMessages)
	r.Post("/users/{id}/messages", PostPrivateMessages)

	root.Mount("/api", r)

	log.Fatal(http.ListenAndServe(":5000", root))
}

func GetMessages(w http.ResponseWriter, r *http.Request) {
	_, ok := r.Context().Value(userID).(idType)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(GroupMessages)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func GetPrivateMessages(w http.ResponseWriter, r *http.Request) {
	id, ok := r.Context().Value(userID).(idType)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	var MyMessages []MessgList
	for k, v := range PrivateMessages {
		if k.Recipient == id {
			MyMessages = append(MyMessages, MessgList{ID: k.Sender, Messages: v})
		}
	}
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(MyMessages)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func PostMessages(w http.ResponseWriter, r *http.Request) {
	id, ok := r.Context().Value(userID).(idType)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	d, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	var m MessgBody
	err = json.Unmarshal(d, &m)
	if err != nil || m.Message == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	GroupMessages = append(GroupMessages, MessgFrom{ID: id, Text: m.Message})
}

func PostPrivateMessages(w http.ResponseWriter, r *http.Request) {
	recip := idType(chi.URLParam(r, "id"))
	id, ok := r.Context().Value(userID).(idType)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	d, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	var m MessgBody
	err = json.Unmarshal(d, &m)

	_, isUser := Users[recip]
	if err != nil || m.Message == "" || !isUser {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	PrivateMessages[FromTo{Sender: id, Recipient: idType(recip)}] = append(PrivateMessages[FromTo{Sender: id, Recipient: idType(recip)}], m.Message)
}

func Register(w http.ResponseWriter, r *http.Request) {
	var creds Credentials
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if _, ok := Users[creds.Username]; ok {
		_, err := w.Write([]byte("Username " + creds.Username + " is already occupied"))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else {
		Users[creds.Username] = creds.Password
		_, err := w.Write([]byte("User " + creds.Username + " is successfully registered!"))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}

func SignIn(w http.ResponseWriter, r *http.Request) {
	var creds Credentials
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	expectedPassword, ok := Users[creds.Username]
	if !ok || expectedPassword != creds.Password {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	claims := &Claims{
		Username:       creds.Username,
		StandardClaims: jwt.StandardClaims{},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:  "token",
		Value: tokenString,
		Path:  "/",
	})
}

func Auth(handler http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie("token")
		switch err {
		case nil:
		case http.ErrNoCookie:
			w.WriteHeader(http.StatusUnauthorized)
			return
		default:
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		tknStr := c.Value

		claims := &Claims{}
		tkn, err := jwt.ParseWithClaims(tknStr, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})
		if err != nil {
			if err == jwt.ErrSignatureInvalid {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if !tkn.Valid {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		idCtx := context.WithValue(r.Context(), userID, claims.Username)

		handler.ServeHTTP(w, r.WithContext(idCtx))
	}

	return http.HandlerFunc(fn)
}
