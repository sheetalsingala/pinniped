// Copyright 2020 the Pinniped contributors. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package oidctestutil

import (
	"context"
	"net/url"

	"golang.org/x/oauth2"

	"go.pinniped.dev/internal/oidc/provider"
	"go.pinniped.dev/pkg/oidcclient/nonce"
	"go.pinniped.dev/pkg/oidcclient/oidctypes"
	"go.pinniped.dev/pkg/oidcclient/pkce"
)

// Test helpers for the OIDC package.

// ExchangeAuthcodeAndValidateTokenArgs is a POGO (plain old go object?) used to spy on calls to
// TestUpstreamOIDCIdentityProvider.ExchangeAuthcodeAndValidateTokensFunc().
type ExchangeAuthcodeAndValidateTokenArgs struct {
	Ctx                  context.Context
	Authcode             string
	PKCECodeVerifier     pkce.Code
	ExpectedIDTokenNonce nonce.Nonce
	RedirectURI          string
}

type TestUpstreamOIDCIdentityProvider struct {
	Name                                  string
	ClientID                              string
	AuthorizationURL                      url.URL
	UsernameClaim                         string
	GroupsClaim                           string
	Scopes                                []string
	ExchangeAuthcodeAndValidateTokensFunc func(
		ctx context.Context,
		authcode string,
		pkceCodeVerifier pkce.Code,
		expectedIDTokenNonce nonce.Nonce,
	) (*oidctypes.Token, error)

	exchangeAuthcodeAndValidateTokensCallCount int
	exchangeAuthcodeAndValidateTokensArgs      []*ExchangeAuthcodeAndValidateTokenArgs
}

func (u *TestUpstreamOIDCIdentityProvider) GetName() string {
	return u.Name
}

func (u *TestUpstreamOIDCIdentityProvider) GetClientID() string {
	return u.ClientID
}

func (u *TestUpstreamOIDCIdentityProvider) GetAuthorizationURL() *url.URL {
	return &u.AuthorizationURL
}

func (u *TestUpstreamOIDCIdentityProvider) GetScopes() []string {
	return u.Scopes
}

func (u *TestUpstreamOIDCIdentityProvider) GetUsernameClaim() string {
	return u.UsernameClaim
}

func (u *TestUpstreamOIDCIdentityProvider) GetGroupsClaim() string {
	return u.GroupsClaim
}

func (u *TestUpstreamOIDCIdentityProvider) ExchangeAuthcodeAndValidateTokens(
	ctx context.Context,
	authcode string,
	pkceCodeVerifier pkce.Code,
	expectedIDTokenNonce nonce.Nonce,
	redirectURI string,
) (*oidctypes.Token, error) {
	if u.exchangeAuthcodeAndValidateTokensArgs == nil {
		u.exchangeAuthcodeAndValidateTokensArgs = make([]*ExchangeAuthcodeAndValidateTokenArgs, 0)
	}
	u.exchangeAuthcodeAndValidateTokensCallCount++
	u.exchangeAuthcodeAndValidateTokensArgs = append(u.exchangeAuthcodeAndValidateTokensArgs, &ExchangeAuthcodeAndValidateTokenArgs{
		Ctx:                  ctx,
		Authcode:             authcode,
		PKCECodeVerifier:     pkceCodeVerifier,
		ExpectedIDTokenNonce: expectedIDTokenNonce,
		RedirectURI:          redirectURI,
	})
	return u.ExchangeAuthcodeAndValidateTokensFunc(ctx, authcode, pkceCodeVerifier, expectedIDTokenNonce)
}

func (u *TestUpstreamOIDCIdentityProvider) ExchangeAuthcodeAndValidateTokensCallCount() int {
	return u.exchangeAuthcodeAndValidateTokensCallCount
}

func (u *TestUpstreamOIDCIdentityProvider) ExchangeAuthcodeAndValidateTokensArgs(call int) *ExchangeAuthcodeAndValidateTokenArgs {
	if u.exchangeAuthcodeAndValidateTokensArgs == nil {
		u.exchangeAuthcodeAndValidateTokensArgs = make([]*ExchangeAuthcodeAndValidateTokenArgs, 0)
	}
	return u.exchangeAuthcodeAndValidateTokensArgs[call]
}

func (u *TestUpstreamOIDCIdentityProvider) ValidateToken(_ context.Context, _ *oauth2.Token, _ nonce.Nonce) (*oidctypes.Token, error) {
	panic("implement me")
}

func NewIDPListGetter(upstreamOIDCIdentityProviders ...*TestUpstreamOIDCIdentityProvider) provider.DynamicUpstreamIDPProvider {
	idpProvider := provider.NewDynamicUpstreamIDPProvider()
	upstreams := make([]provider.UpstreamOIDCIdentityProviderI, len(upstreamOIDCIdentityProviders))
	for i := range upstreamOIDCIdentityProviders {
		upstreams[i] = provider.UpstreamOIDCIdentityProviderI(upstreamOIDCIdentityProviders[i])
	}
	idpProvider.SetIDPList(upstreams)
	return idpProvider
}

// Declare a separate type from the production code to ensure that the state param's contents was serialized
// in the format that we expect, with the json keys that we expect, etc. This also ensure that the order of
// the serialized fields is the same, which doesn't really matter expect that we can make simpler equality
// assertions about the redirect URL in this test.
type ExpectedUpstreamStateParamFormat struct {
	P string `json:"p"`
	U string `json:"u"`
	N string `json:"n"`
	C string `json:"c"`
	K string `json:"k"`
	V string `json:"v"`
}
