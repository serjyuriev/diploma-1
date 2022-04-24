package service

import (
	"database/sql"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/rs/zerolog"
	"github.com/serjyuriev/diploma-1/internal/pkg/config"
	"github.com/serjyuriev/diploma-1/internal/pkg/mocks"
	"github.com/serjyuriev/diploma-1/internal/pkg/models"
	"github.com/stretchr/testify/require"
)

func Test_CreateNewOrder(t *testing.T) {
	ctrl := gomock.NewController(t)

	ma := mocks.NewMockAccrual(ctrl)

	// Asserts that the first and only call to Bar() is passed 99.
	// Anything else will fail.
	first := ma.EXPECT().GetOrderStatus(nil, gomock.Eq("6122")).Return(
		&models.Order{
			Number:        "6122",
			AccrualStatus: "REGISTERED",
		}, nil,
	)
	second := ma.EXPECT().GetOrderStatus(nil, gomock.Eq("6122")).Return(
		&models.Order{
			Number:        "6122",
			AccrualStatus: "PROCESSING",
		}, nil,
	)
	third := ma.EXPECT().GetOrderStatus(nil, gomock.Eq("6122")).Return(
		&models.Order{
			Number:        "6122",
			AccrualStatus: "PROCESSED",
			Accrual:       30012,
		}, nil,
	)
	gomock.InOrder(
		first,
		second,
		third,
	)

	mr := mocks.NewMockRepository(ctrl)
	mr.EXPECT().SelectOrderByNumber(nil, gomock.Eq("6122")).Return(nil, sql.ErrNoRows)
	mr.EXPECT().UpdateOrderStatus(nil, gomock.Eq("6122"), gomock.Any()).MinTimes(2)
	mr.EXPECT().InsertOrder(nil, gomock.Eq("6122"), gomock.Eq(2)).Return(int64(2), nil)
	mr.EXPECT().InsertAccrual(nil, gomock.Eq(2), gomock.Eq(300.12), gomock.Eq(int64(2))).Return(nil)

	output := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: "02-01-2006 15:04:05 MST",
	}
	logger := zerolog.New(output).With().Timestamp().Logger()

	svc := &service{
		config: config.Config{
			DatabaseURI:               "postgres://gopher:G0ph3R@localhost:5432/gophermart",
			AccrualSystemSurveyPeriod: 5,
		},
		logger:  logger,
		accrual: ma,
		repo:    mr,
		jobChan: make(chan *job),
	}
	svc.polling(5)
	err := svc.CreateNewOrder(nil, "6122", 2)
	require.NoError(t, err)
}
