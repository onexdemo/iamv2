// Package jwt implements the iamctl jwt commands.
package jwt

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v4"
	"github.com/spf13/cobra"
)

// ArgList is a repeatable key=value flag.
type ArgList map[string]string

// String renders the map as JSON.
func (l ArgList) String() string {
	data, _ := json.Marshal(l)
	return string(data)
}

// Set adds a key=value pair.
func (l ArgList) Set(arg string) error {
	key, value, ok := strings.Cut(arg, "=")
	if !ok || key == "" {
		return fmt.Errorf("invalid argument %q: expected key=value", arg)
	}

	l[key] = value

	return nil
}

// Type returns the pflag type name.
func (ArgList) Type() string { return "map" }

// NewCmdJWT creates the jwt command group.
func NewCmdJWT() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "jwt",
		Short: "Sign, show and verify JWT tokens",
		Run: func(cmd *cobra.Command, _ []string) {
			_ = cmd.Help()
		},
	}

	cmd.AddCommand(newCmdSign(), newCmdShow(), newCmdVerify())

	return cmd
}

func newCmdSign() *cobra.Command {
	var (
		timeout   time.Duration
		notBefore time.Duration
		algorithm string
		audience  string
		issuer    string
		claims    = ArgList{}
		headers   = ArgList{}
	)

	cmd := &cobra.Command{
		Use:   "sign SECRETID SECRETKEY",
		Short: "Sign a JWT token with an access key pair",
		Args:  cobra.ExactArgs(2),
		Example: `  # Sign a token valid for two hours
  iamctl jwt sign tgydj8d9EQSnFqKf iBdEdFNBLN1nR3fV --timeout=2h --algorithm=HS256`,
		RunE: func(cmd *cobra.Command, args []string) error {
			switch algorithm {
			case "HS256", "HS384", "HS512":
			default:
				return fmt.Errorf("unsupported signing algorithm %q", algorithm)
			}

			mapClaims := jwtlib.MapClaims{
				"exp": time.Now().Add(timeout).Unix(),
				"iat": time.Now().Unix(),
				"nbf": time.Now().Add(notBefore).Unix(),
				"aud": audience,
				"iss": issuer,
			}
			for key, value := range claims {
				mapClaims[key] = value
			}

			token := jwtlib.NewWithClaims(jwtlib.GetSigningMethod(algorithm), mapClaims)
			for key, value := range headers {
				token.Header[key] = value
			}

			token.Header["kid"] = args[0]

			signed, err := token.SignedString([]byte(args[1]))
			if err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), signed)

			return nil
		},
	}

	cmd.Flags().DurationVar(&timeout, "timeout", 2*time.Hour, "JWT token expiration.")
	cmd.Flags().DurationVar(&notBefore, "not-before", 0, "Time before which the JWT must not be accepted.")
	cmd.Flags().StringVar(&algorithm, "algorithm", "HS256", "Signing algorithm: HS256, HS384 or HS512.")
	cmd.Flags().StringVar(&audience, "audience", "iam.authz.marmotedu.com", "JWT audience.")
	cmd.Flags().StringVar(&issuer, "issuer", "iamctl", "JWT issuer.")
	cmd.Flags().Var(&claims, "claim", "Add an additional claim (key=value). Can be repeated.")
	cmd.Flags().Var(&headers, "header", "Add an additional header (key=value). Can be repeated.")

	return cmd
}

func newCmdShow() *cobra.Command {
	return &cobra.Command{
		Use:   "show TOKEN",
		Short: "Show the header and claims of a JWT token without verifying it",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			parser := jwtlib.NewParser()
			token, _, err := parser.ParseUnverified(args[0], jwtlib.MapClaims{})
			if err != nil {
				return err
			}

			header, _ := json.MarshalIndent(token.Header, "", "  ")
			claims, _ := json.MarshalIndent(token.Claims, "", "  ")

			fmt.Fprintf(cmd.OutOrStdout(), "Header:\n%s\n\nClaims:\n%s\n", header, claims)

			return nil
		},
	}
}

func newCmdVerify() *cobra.Command {
	return &cobra.Command{
		Use:   "verify SECRETKEY TOKEN",
		Short: "Verify a JWT token signature",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			token, err := jwtlib.Parse(args[1], func(token *jwtlib.Token) (any, error) {
				if _, ok := token.Method.(*jwtlib.SigningMethodHMAC); !ok {
					return nil, errors.New("unexpected signing method")
				}

				return []byte(args[0]), nil
			})
			if err != nil {
				return err
			}

			if !token.Valid {
				return errors.New("token is invalid")
			}

			fmt.Fprintln(cmd.OutOrStdout(), "token is valid")

			return nil
		},
	}
}
