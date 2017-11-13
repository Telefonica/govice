/**
 * @license
 * Copyright 2017 Telefónica Investigación y Desarrollo, S.A.U
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/Telefonica/govice"
)

// User type.
type User struct {
	Login *string `json:"login"`
	Name  *string `json:"name"`
	Pin   *int    `json:"pin"`
}

// UsersService type.
type UsersService struct {
	v  *govice.Validator
	db map[string]*User
}

// NewUsersService creates a new instance of UserService.
func NewUsersService(v *govice.Validator) *UsersService {
	return &UsersService{
		v:  v,
		db: make(map[string]*User),
	}
}

// CreateUser registers the user in the map.
func (u *UsersService) CreateUser(w http.ResponseWriter, r *http.Request) {
	var user User
	if err := u.v.ValidateRequestBody("user", r, &user); err != nil {
		govice.ReplyWithError(w, r, err)
		return
	}
	login := *user.Login
	u.db[login] = &user
	w.Header().Add("Location", "/users/"+login)
	w.WriteHeader(http.StatusCreated)
}

// GetUser queries a user in the map.
func (u *UsersService) GetUser(w http.ResponseWriter, r *http.Request) {
	login := mux.Vars(r)["login"]
	user, ok := u.db[login]
	if !ok {
		govice.ReplyWithError(w, r, govice.NotFoundError)
		return
	}
	response, err := json.Marshal(user)
	if err != nil {
		e := govice.NewServerError(fmt.Sprintf("error marshalling the user. %s", err))
		govice.ReplyWithError(w, r, e)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}

// DeleteUser deletes a user from the map.
func (u *UsersService) DeleteUser(w http.ResponseWriter, r *http.Request) {
	login := mux.Vars(r)["login"]
	_, ok := u.db[login]
	if !ok {
		govice.ReplyWithError(w, r, govice.NotFoundError)
		return
	}
	delete(u.db, login)
	w.WriteHeader(http.StatusNoContent)
}
