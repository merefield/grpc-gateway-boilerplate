/*
Package sdk is the gRPC implementation of the SDK gRPC server
Copyright 2018 Portworx

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package iam

import (
	"context"

	"github.com/go-pg/pg/orm"
	"github.com/merefield/grpc-user-api/internal/model"
	"github.com/merefield/grpc-user-api/proto/iam"
	"github.com/twitchtv/twirp"
)

// IdentityServer is an implementation of the gRPC OpenStorageIdentityServer interface
type IAMServer struct {
	server serverAccessor
}

// TokenGenerator generates new jwt token
type TokenGenerator interface {
	GenerateToken(*model.AuthUser) (string, error)
}

// UserDB represents user database interface
type UserDB interface {
	FindByAuth(orm.DB, string) (*model.User, error)
	FindByToken(orm.DB, string) (*model.User, error)
	UpdateLastLogin(orm.DB, *model.User) error
}

// Securer represents password securing service
type Securer interface {
	MatchesHash(string, string) bool
}

var (
	invalidUserPW = twirp.NewError(twirp.PermissionDenied, "invalid username or password")
	invalidToken  = twirp.NewError(twirp.PermissionDenied, "invalid token")
)

// Auth tries to authenticate user given username and password
func (s *IAMServer) Auth(c context.Context, req *iam.AuthReq) (*iam.AuthResp, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	dbCtx := s.db.WithContext(c)

	usr, err := s.udb.FindByAuth(dbCtx, req.Auth)
	if err != nil {
		return nil, invalidUserPW
	}

	if !s.sec.MatchesHash(usr.Password, req.Password) {
		return nil, invalidUserPW
	}

	token, err := s.tg.GenerateToken(&model.AuthUser{
		Id:       usr.Id,
		TenantId: usr.TenantId,
		Username: usr.Username,
		Email:    usr.Email,
		Role:     model.AccessRole(usr.RoleId),
	})

	if err != nil {
		return nil, err
	}

	uToken := xid.New().String()

	usr.UpdateLoginDetails(uToken)

	if err = s.udb.UpdateLastLogin(dbCtx, usr); err != nil {
		return nil, err
	}

	return &iam.AuthResp{
		Token:        token,
		RefreshToken: uToken,
	}, nil
}

// Refresh refreshes user's jwt token
func (s *IAMServer) Refresh(c context.Context, req *iam.RefreshReq) (*iam.RefreshResp, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	usr, err := s.udb.FindByToken(s.db.WithContext(c), req.Token)
	if err != nil {
		return nil, invalidToken
	}

	token, err := s.tg.GenerateToken(&model.AuthUser{
		Id:       usr.Id,
		TenantId: usr.TenantId,
		Username: usr.Username,
		Email:    usr.Email,
		Role:     model.AccessRole(usr.RoleId),
	})

	if err != nil {
		return nil, err
	}

	return &iam.RefreshResp{
		Token: token,
	}, nil
}
