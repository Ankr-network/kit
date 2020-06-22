package auth

import (
	"context"
	"crypto/rsa"
	"errors"
	"github.com/dgrijalva/jwt-go"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io/ioutil"
	"regexp"
)

var (
	ErrExpired            = errors.New("token expired")
	ErrMissingMetadata    = errors.New("miss metadata")
	ErrEmptyAuthorization = errors.New("empty authorization header")
	ErrInvalidClaim       = errors.New("invalid jwt claim")
	ErrInvalidToken       = errors.New("invalid token")
)

type Verifier interface {
	GetKey() []byte
	Verify(tokenString string) (*jwt.Token, error)
	GRPCUnaryInterceptor() grpc.UnaryServerInterceptor
	VerifyContext(ctx context.Context) (context.Context, error)
}

type VerifierOptions struct {
	ExcludeMethods []string
	TokenBlacklist Blacklist
}

type VerifierOption func(opts *VerifierOptions)

func ExcludeMethods(method ...string) VerifierOption {
	return func(opts *VerifierOptions) {
		opts.ExcludeMethods = append(opts.ExcludeMethods, method...)
	}
}

func TokenBlacklist(bl Blacklist) VerifierOption {
	return func(opts *VerifierOptions) {
		opts.TokenBlacklist = bl
	}
}

func NewVerifier(publicKeyPath string, opts ...VerifierOption) (Verifier, error) {
	rawKey, err := ioutil.ReadFile(publicKeyPath)
	if err != nil {
		return nil, err
	}

	key, err := jwt.ParseRSAPublicKeyFromPEM(rawKey)

	// default
	options := &VerifierOptions{
		ExcludeMethods: []string{`/.+Internal.+/.+`, `/grpc\.health\.v1\.Health/Check`},
	}

	for _, o := range opts {
		o(options)
	}

	excludePatterns := make([]*regexp.Regexp, len(options.ExcludeMethods))
	for i, m := range options.ExcludeMethods {
		r, err := regexp.Compile(m)
		if err != nil {
			log.Error("invalid method pattern error", zap.Error(err))
			return nil, err
		}
		excludePatterns[i] = r
	}

	return &verifier{
		rawKey:          rawKey,
		key:             key,
		excludePatterns: excludePatterns,
		blacklist:       options.TokenBlacklist,
	}, nil
}

type verifier struct {
	rawKey          []byte
	key             *rsa.PublicKey
	excludePatterns []*regexp.Regexp
	blacklist       Blacklist
}

func (p *verifier) Verify(tokenString string) (*jwt.Token, error) {
	if p.blacklist != nil {
		if err := p.blacklist.CheckAccess(tokenString); err != nil {
			if errors.Is(err, ErrExpiredAccess) {
				return nil, ErrInvalidToken
			}
			return nil, err
		}
	}
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return p.key, nil
	})
	if vErr, ok := err.(*jwt.ValidationError); ok {
		if vErr.Errors&jwt.ValidationErrorExpired == jwt.ValidationErrorExpired {
			return token, ErrExpired
		}
	}
	return token, err
}

func (p *verifier) GetKey() []byte {
	return p.rawKey
}

func (p *verifier) GRPCUnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		// exclude methods
		if p.matchMethod(info.FullMethod) {
			return handler(ctx, req)
		}

		outCtx, err := p.VerifyContext(ctx)
		if err != nil {
			return nil, status.Errorf(codes.Unauthenticated, "AuthError:%v", err)
		}
		return handler(outCtx, req)
	}
}

func (p *verifier) VerifyContext(ctx context.Context) (context.Context, error) {
	tokenString, err := ExtractToken(ctx)
	if err != nil {
		return nil, err
	}

	token, err := p.Verify(tokenString)
	if err != nil {
		return nil, ErrInvalidToken
	}

	claim, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		log.Info("invalid claim", zap.Reflect("claims", token.Claims))
		return nil, ErrInvalidClaim
	}

	return context.WithValue(ctx, "claim", claim), nil
}

func (p *verifier) matchMethod(method string) bool {
	for _, r := range p.excludePatterns {
		if r.MatchString(method) {
			return true
		}
	}
	return false
}
