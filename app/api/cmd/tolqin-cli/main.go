package main

import (
	"errors"
	"log"

	_ "github.com/lib/pq"
	"github.com/ztimes2/tolqin/app/api/internal/auth"
	authpsql "github.com/ztimes2/tolqin/app/api/internal/auth/psql"
	"github.com/ztimes2/tolqin/app/api/internal/cmd"
	"github.com/ztimes2/tolqin/app/api/internal/pkg/psqlutil"
	"github.com/ztimes2/tolqin/app/api/internal/pkg/valerra"
	"github.com/ztimes2/tolqin/app/api/internal/service/operator"
)

func main() {
	db, err := psqlutil.NewDB(psqlutil.DriverNamePQ, psqlutil.Config{
		Host:         "localhost",
		Port:         "5432",
		Username:     "root",
		Password:     "root",
		DatabaseName: "tolqin",
		SSLMode:      psqlutil.SSLModeDisabled,
	})
	if err != nil {
		log.Fatalf("❌ %s", err.Error())
	}

	s := operator.NewService(
		auth.NewPasswordSalter(),
		auth.NewPasswordHasher(),
		authpsql.NewUserStore(db),
	)

	if err := cmd.New(s).Execute(); err != nil {
		var vErr *valerra.Errors
		if errors.As(err, &vErr) {
			log.Fatalf("❌ %s", vErr.Errors())
		}
		log.Fatalf("❌ %s", err.Error())
	}
}
