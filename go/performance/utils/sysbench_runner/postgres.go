package sysbench_runner

import (
	"context"
	"database/sql"
	"fmt"
	"syscall"

	_ "github.com/lib/pq"
)

const (
	postgresInitDbDataDirFlag         = "--pgdata"
	postgresUsernameFlag              = "--username"
	postgresUsername                  = "postgres"
	postgresDataDirFlag               = "-D"
	postgresDropDatabaseSqlTemplate   = "DROP DATABASE IF EXISTS %s;"
	postgresDropUserSqlTemplate       = "DROP USER IF EXISTS %s;"
	postgresCreateUserSqlTemplate     = "CREATE USER %s WITH PASSWORD '%s';"
	postgresCreateDatabaseSqlTemplate = "CREATE DATABASE %s WITH OWNER %s;"
)

type postgresBenchmarkerImpl struct {
	dir          string // cwd
	config       *Config
	serverConfig *ServerConfig
}

var _ Benchmarker = &postgresBenchmarkerImpl{}

func NewPostgresBenchmarker(dir string, config *Config, serverConfig *ServerConfig) *postgresBenchmarkerImpl {
	return &postgresBenchmarkerImpl{
		dir:          dir,
		config:       config,
		serverConfig: serverConfig,
	}
}

func (b *postgresBenchmarkerImpl) initDataDir(ctx context.Context) (string, error) {
	serverDir, err := CreateServerDir(dbName)
	if err != nil {
		return "", err
	}

	pgInit := ExecCommand(ctx, b.serverConfig.InitExec, fmt.Sprintf("%s=%s", postgresInitDbDataDirFlag, serverDir), fmt.Sprintf("%s=%s", postgresUsernameFlag, postgresUsername))
	err = pgInit.Run()
	if err != nil {
		return "", err
	}

	return serverDir, nil
}

func (b *postgresBenchmarkerImpl) createTestingDb(ctx context.Context) (err error) {
	psqlconn := fmt.Sprintf(psqlDsnTemplate, b.serverConfig.Host, b.serverConfig.Port, postgresUsername, "", dbName)

	var db *sql.DB
	db, err = sql.Open("postgres", psqlconn)
	if err != nil {
		return
	}
	defer func() {
		rerr := db.Close()
		if err == nil {
			err = rerr
		}
	}()
	err = db.PingContext(ctx)
	if err != nil {
		return
	}

	stmts := []string{
		fmt.Sprintf(postgresDropDatabaseSqlTemplate, dbName),
		fmt.Sprintf(postgresDropUserSqlTemplate, sysbenchUsername),
		fmt.Sprintf(postgresCreateUserSqlTemplate, sysbenchUsername, sysbenchPassLocal),
		fmt.Sprintf(postgresCreateDatabaseSqlTemplate, dbName, sysbenchUsername),
	}

	for _, s := range stmts {
		_, err = db.ExecContext(ctx, s)
		if err != nil {
			return
		}
	}

	return
}

func (b *postgresBenchmarkerImpl) Benchmark(ctx context.Context) (Results, error) {
	serverDir, err := b.initDataDir(ctx)
	if err != nil {
		return nil, err
	}

	serverParams, err := b.serverConfig.GetServerArgs()
	if err != nil {
		return nil, err
	}
	serverParams = append(serverParams, postgresDataDirFlag, serverDir)

	server := NewServer(ctx, serverDir, b.serverConfig, syscall.SIGTERM, serverParams)
	err = server.Start(ctx)
	if err != nil {
		return nil, err
	}

	err = b.createTestingDb(ctx)
	if err != nil {
		return nil, err
	}

	tests, err := GetTests(b.config, b.serverConfig, nil)
	if err != nil {
		return nil, err
	}

	results := make(Results, 0)
	for i := 0; i < b.config.Runs; i++ {
		for _, test := range tests {
			tester := NewSysbenchTester(b.config, b.serverConfig, test, stampFunc)
			r, err := tester.Test(ctx)
			if err != nil {
				server.Stop(ctx)
				return nil, err
			}
			results = append(results, r)
		}
	}

	err = server.Stop(ctx)
	if err != nil {
		return nil, err
	}

	return results, nil
}
