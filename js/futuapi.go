package js

import (
	futuapi "github.com/CCLooMi/go-futu-api"
	"github.com/futuopen/ftapi4go/pb/getglobalstate"
	"github.com/futuopen/ftapi4go/pb/qotcommon"
	"github.com/futuopen/ftapi4go/pb/qotgetmarketstate"
	"github.com/futuopen/ftapi4go/pb/qotgetsubinfo"
	"github.com/futuopen/ftapi4go/pb/qotgetusersecuritygroup"
	"github.com/futuopen/ftapi4go/pb/qotrequesthistorykl"
	"github.com/futuopen/ftapi4go/pb/trdcommon"
	"github.com/gogo/protobuf/proto"
	"go.uber.org/fx"
	"golang.org/x/net/context"
	"log"
	"math"
	"time"
	"wios_server/conf"
)

type FTApi struct {
	fapi *futuapi.FutuAPI
	conf *conf.FutuApiConfig
}

func NewFTApi(fapi *futuapi.FutuAPI, conf *conf.FutuApiConfig) *FTApi {
	return &FTApi{fapi, conf}
}
func (f *FTApi) ConnID() uint64 {
	return f.fapi.ConnID()
}
func (f *FTApi) GetGlobalState(ctx context.Context) (*getglobalstate.S2C, error) {
	return f.fapi.GetGlobalState(ctx)
}
func (f *FTApi) IsConnected(ctx context.Context) bool {
	if _, err := f.fapi.GetGlobalState(ctx); err != nil {
		return false
	}
	return true
}
func (f *FTApi) Connect(ctx context.Context) error {
	if f.IsConnected(ctx) {
		return nil
	}
	return f.fapi.Connect(ctx, f.conf.ApiAddr)
}
func (f *FTApi) GetAccList(ctx context.Context,
	category trdcommon.TrdCategory, generalAcc bool) ([]*trdcommon.TrdAcc, error) {
	return f.fapi.GetAccList(ctx, category,
		&futuapi.OptionalBool{Value: generalAcc})
}
func (f *FTApi) GetFunds(ctx context.Context, acc *trdcommon.TrdAcc, refresh bool) (*trdcommon.Funds, error) {
	ma := acc.TrdMarketAuthList[0]
	return f.fapi.GetFunds(ctx,
		&trdcommon.TrdHeader{
			TrdEnv:    acc.TrdEnv,
			TrdMarket: proto.Int32(ma),
			AccID:     acc.AccID,
		}, &futuapi.OptionalBool{
			Value: refresh,
		}, trdcommon.Currency(ma))
}
func (f *FTApi) GetPositions(ctx context.Context, acc *trdcommon.TrdAcc, refresh bool) ([]*trdcommon.Position, error) {
	return f.fapi.GetPositionList(ctx,
		&trdcommon.TrdHeader{
			TrdEnv:    acc.TrdEnv,
			TrdMarket: proto.Int32(acc.TrdMarketAuthList[0]),
			AccID:     acc.AccID,
		}, &trdcommon.TrdFilterConditions{},
		&futuapi.OptionalDouble{Value: -math.MaxFloat64},
		&futuapi.OptionalDouble{Value: math.MaxFloat64},
		&futuapi.OptionalBool{Value: refresh})
}
func (f *FTApi) QuerySubs(ctx context.Context, isAll bool) (*qotgetsubinfo.S2C, error) {
	return f.fapi.QuerySubscription(ctx, isAll)
}
func (f *FTApi) GetSecGroup(ctx context.Context, groupType qotgetusersecuritygroup.GroupType) ([]*qotgetusersecuritygroup.GroupData, error) {
	return f.fapi.GetUserSecurityGroup(ctx, groupType)
}
func (f *FTApi) GetGroupSec(ctx context.Context, group string) ([]*qotcommon.SecurityStaticInfo, error) {
	return f.fapi.GetUserSecurity(ctx, group)
}
func (f *FTApi) GetHistoryKline(ctx context.Context, market int32, code string,
	begin string, end string, klType qotcommon.KLType, rehabType qotcommon.RehabType,
	extTime bool) (*qotrequesthistorykl.S2C, error) {
	sec := &qotcommon.Security{
		Market: proto.Int32(market),
		Code:   proto.String(code),
	}
	//sub first
	if err := f.fapi.Subscribe(ctx,
		[]*qotcommon.Security{sec},
		nil,
		false,
		false,
		false,
		false); err != nil {
		return nil, err
	}
	//final unsubscribe
	defer func() {
		f.fapi.Unsubscribe(ctx,
			[]*qotcommon.Security{sec},
			[]qotcommon.SubType{qotcommon.SubType(klType)})
	}()
	return f.fapi.RequestHistoryKLine(ctx,
		sec,
		begin,
		end,
		klType, rehabType,
		&futuapi.OptionalInt32{Value: math.MaxInt32},
		qotcommon.KLFields_KLFields_None, nil,
		&futuapi.OptionalBool{Value: extTime},
	)
}

func (f *FTApi) GetMarketState(ctx context.Context, codes ...string) ([]*qotgetmarketstate.MarketInfo, error) {
	secs := make([]*qotcommon.Security, 0, len(codes))
	for _, code := range codes {
		secs = append(secs, &qotcommon.Security{
			Market: proto.Int32(0),
			Code:   proto.String(code),
		})
	}
	return f.fapi.GetMarketState(ctx, secs)
}
func (f *FTApi) TrdMarketName(i int32) string {
	return trdcommon.TrdMarket_name[i]
}
func (f *FTApi) TrdEnvName(i int32) string {
	return trdcommon.TrdEnv_name[i]
}
func (f *FTApi) TrdCategoryName(i int32) string {
	return trdcommon.TrdCategory_name[i]
}
func (f *FTApi) TrdSecMarketName(i int32) string {
	return trdcommon.TrdSecMarket_name[i]
}
func (f *FTApi) TrdSideName(i int32) string {
	return trdcommon.TrdSide_name[i]
}
func (f *FTApi) OrderTypeName(i int32) string {
	return trdcommon.OrderType_name[i]
}
func (f *FTApi) TrailTypeName(i int32) string {
	return trdcommon.TrailType_name[i]
}
func (f *FTApi) OrderStatusName(i int32) string {
	return trdcommon.OrderStatus_name[i]
}
func (f *FTApi) OrderFillStatusName(i int32) string {
	return trdcommon.OrderFillStatus_name[i]
}
func (f *FTApi) PositionSideName(i int32) string {
	return trdcommon.PositionSide_name[i]
}
func (f *FTApi) TrdAccTypeName(i int32) string {
	return trdcommon.TrdAccType_name[i]
}
func (f *FTApi) TrdAccStatusName(i int32) string {
	return trdcommon.TrdAccStatus_name[i]
}
func (f *FTApi) CurrencyName(i int32) string {
	return trdcommon.Currency_name[i]
}
func (f *FTApi) CltRiskLevelName(i int32) string {
	return trdcommon.CltRiskLevel_name[i]
}
func (f *FTApi) TimeInForceName(i int32) string {
	return trdcommon.TimeInForce_name[i]
}
func (f *FTApi) SecurityFirmName(i int32) string {
	return trdcommon.SecurityFirm_name[i]
}
func (f *FTApi) SimAccTypeName(i int32) string {
	return trdcommon.SimAccType_name[i]
}
func (f *FTApi) CltRiskStatusName(i int32) string {
	return trdcommon.CltRiskStatus_name[i]
}
func (f *FTApi) DTStatusName(i int32) string {
	return trdcommon.DTStatus_name[i]
}
func (f *FTApi) ModifyOrderOpName(i int32) string {
	return trdcommon.ModifyOrderOp_name[i]
}
func newFutuApi(config *conf.Config) *futuapi.FutuAPI {
	api := futuapi.NewFutuAPI()
	api.SetClientInfo(config.DHTConf.PeerId, 1)
	return api
}
func connectFutuApi(lc fx.Lifecycle, api *futuapi.FutuAPI, config *conf.Config) *futuapi.FutuAPI {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				for {
					err := api.Connect(context.Background(), config.FutuApiConf.ApiAddr)
					if err != nil {
						time.Sleep(10 * time.Second)
						continue
					}
					log.Println("Successfully connected to Futu API.")
					break
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			go func() {
				if err := api.Close(ctx); err != nil {
					log.Printf("Failed to close Futu API: %v", err)
				}
				log.Println("Successfully closed Futu API.")
			}()
			return nil
		},
	})
	return api
}

var futuApiModule = fx.Options(
	fx.Provide(newFutuApi),
	fx.Invoke(
		connectFutuApi,
		func(fapi *futuapi.FutuAPI, config *conf.Config) {
			RegExport("futuapi", NewFTApi(fapi, &config.FutuApiConf))
		},
	),
)
