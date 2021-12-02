package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/MicahParks/keyfunc"
	"github.com/go-chi/chi"
	"github.com/go-chi/cors"
	"github.com/golang-jwt/jwt/v4"
)

var userIDKey = struct{}{}

type UserContext struct {
	UserID string   `json:"userId"`
	Roles  []string `json:"roles"`
}

func checkAuth(jwksURL string) (func(http.Handler) http.Handler, error) {
	jwks, err := keyfunc.Get(jwksURL, keyfunc.Options{})
	if err != nil {
		log.Printf("Failed to get the JWKs from the given URL.\nError:%s\n", err.Error())
		return nil, err
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h := r.Header.Get("Authorization")
			parts := strings.Split(h, " ")
			if len(parts) != 2 {
				http.Error(w, "403", 403)
				return
			}

			token, err := jwt.Parse(parts[1], jwks.Keyfunc)
			if err != nil {
				log.Printf("err: %s", err)
				http.Error(w, "403", 403)
				return
			}
			if !token.Valid {
				log.Printf("!token.Valid")
				http.Error(w, "403", 403)
				return
			}
			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				log.Printf("!claims")
				http.Error(w, "403", 403)
				return
			}
			userID, ok := claims["sub"].(string)
			if !ok {
				log.Printf("!sub")
				http.Error(w, "403", 403)
				return
			}
			realmAccess, ok := claims["realm_access"].(map[string]interface{})
			if !ok {
				log.Printf("!claims: %v", claims["realm_access"])
				http.Error(w, "403", 403)
				return
			}
			iroles, ok := realmAccess["roles"].([]interface{})
			if !ok {
				log.Printf("!roles: %t", realmAccess["roles"])
				http.Error(w, "403", 403)
				return
			}

			roles := make([]string, 0)
			for _, irole := range iroles {
				role, ok := irole.(string)
				if !ok {
					log.Printf("!roles: %t", realmAccess["roles"])
					http.Error(w, "403", 403)
					return
				}
				roles = append(roles, role)
			}
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), userIDKey, &UserContext{
				UserID: userID,
				Roles:  roles,
			})))
		})
	}, nil

}

func main() {
	r := chi.NewRouter()
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: true,
	}))

	jwksURL := "http://localhost:8080/auth/realms/alibox/protocol/openid-connect/certs"
	checkAuthMiddleware, err := checkAuth(jwksURL)
	if err != nil {
		log.Fatalf("err: %s", err)
	}
	r.Use(checkAuthMiddleware)
	r.Get("/api/v1/getPageData", func(w http.ResponseWriter, r *http.Request) {
		uCtx, ok := r.Context().Value(userIDKey).(*UserContext)
		if !ok {
			http.Error(w, "403", 403)
			return
		}
		json.NewEncoder(w).Encode(uCtx)
	})

	http.ListenAndServe(":11488", r)
}
