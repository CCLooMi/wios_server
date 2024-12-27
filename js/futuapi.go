package js

import (
	"fmt"
	futuapi "github.com/CCLooMi/go-futu-api"
	"github.com/futuopen/ftapi4go/pb/common"
	"github.com/futuopen/ftapi4go/pb/getglobalstate"
	"github.com/futuopen/ftapi4go/pb/qotcommon"
	"github.com/futuopen/ftapi4go/pb/qotgetmarketstate"
	"github.com/futuopen/ftapi4go/pb/qotgetsubinfo"
	"github.com/futuopen/ftapi4go/pb/qotgetusersecuritygroup"
	"github.com/futuopen/ftapi4go/pb/qotrequesthistorykl"
	"github.com/futuopen/ftapi4go/pb/qotsetpricereminder"
	"github.com/futuopen/ftapi4go/pb/trdcommon"
	"github.com/futuopen/ftapi4go/pb/trdplaceorder"
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
func (f *FTApi) GetAccList(ctx context.Context, category trdcommon.TrdCategory, generalAcc bool) ([]*trdcommon.TrdAcc, error) {
	return f.fapi.GetAccList(ctx, category,
		&futuapi.OptionalBool{Value: generalAcc})
}
func (f *FTApi) GetCNSimAcc(ctx context.Context, generalAcc bool) (*trdcommon.TrdAcc, error) {
	a, err := f.GetAccList(ctx, trdcommon.TrdCategory_TrdCategory_Security, generalAcc)
	if err != nil {
		return nil, err
	}
	for _, v := range a {
		if *v.TrdEnv == int32(trdcommon.TrdEnv_TrdEnv_Simulate) {
			if v.TrdMarketAuthList[0] == int32(trdcommon.TrdMarket_TrdMarket_CN) {
				return v, nil
			}
		}
	}
	return nil, nil
}
func (f *FTApi) GetUSSimAcc(ctx context.Context, generalAcc bool) (*trdcommon.TrdAcc, error) {
	a, err := f.GetAccList(ctx, trdcommon.TrdCategory_TrdCategory_Security, generalAcc)
	if err != nil {
		return nil, err
	}
	for _, v := range a {
		if *v.TrdEnv == int32(trdcommon.TrdEnv_TrdEnv_Simulate) {
			if v.TrdMarketAuthList[0] == int32(trdcommon.TrdMarket_TrdMarket_US) {
				return v, nil
			}
		}
	}
	return nil, nil
}
func (f *FTApi) GetHKSimAcc(ctx context.Context, generalAcc bool) (*trdcommon.TrdAcc, error) {
	a, err := f.GetAccList(ctx, trdcommon.TrdCategory_TrdCategory_Security, generalAcc)
	if err != nil {
		return nil, err
	}
	for _, v := range a {
		if *v.TrdEnv == int32(trdcommon.TrdEnv_TrdEnv_Simulate) {
			if v.TrdMarketAuthList[0] == int32(trdcommon.TrdMarket_TrdMarket_HK) {
				return v, nil
			}
		}
	}
	return nil, nil
}
func getMktCurrency(mkt int32) trdcommon.Currency {
	switch mkt {
	case 1, 4, 10, 113:
		return trdcommon.Currency_Currency_HKD
	case 2, 11, 123:
		return trdcommon.Currency_Currency_USD
	case 3, 5, 31, 32:
		return trdcommon.Currency_Currency_CNH
	case 6, 12, 41:
		return trdcommon.Currency_Currency_SGD
	case 13, 51:
		return trdcommon.Currency_Currency_JPY
	default:
		return trdcommon.Currency_Currency_Unknown
	}
}
func (f *FTApi) GetFunds(ctx context.Context, acc *trdcommon.TrdAcc, refresh bool) (*trdcommon.Funds, error) {
	mkt := acc.TrdMarketAuthList[0]
	return f.fapi.GetFunds(ctx,
		&trdcommon.TrdHeader{
			TrdEnv:    acc.TrdEnv,
			TrdMarket: proto.Int32(mkt),
			AccID:     acc.AccID,
		}, &futuapi.OptionalBool{
			Value: refresh,
		}, getMktCurrency(mkt))
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
func (f *FTApi) GetSec(ctx context.Context, code string) (*qotcommon.SecurityStaticBasic, error) {
	ss, err := f.GetGroupSec(ctx, "全部")
	if err != nil {
		return nil, err
	}
	for _, v := range ss {
		if *v.Basic.Name == code || *v.Basic.Security.Code == code {
			return v.Basic, nil
		}
	}
	return nil, nil
}
func (f *FTApi) SubSec(ctx context.Context, sec *qotcommon.Security, subType qotcommon.SubType) (func(), error) {
	err := f.fapi.
		Subscribe(ctx, []*qotcommon.Security{sec}, []qotcommon.SubType{subType},
			true, true, false, false)
	if err != nil {
		return nil, err
	}
	return func() {
		t, ca := context.WithTimeout(ctx, 30*time.Second)
		defer ca()
		err := f.fapi.Unsubscribe(t,
			[]*qotcommon.Security{sec},
			[]qotcommon.SubType{subType})
		if err != nil {
			fmt.Printf("Unsubscribe failed: %v\n", err)
			return
		}
	}, nil
}
func klTypeToSubType(klType qotcommon.KLType) qotcommon.SubType {
	switch klType {
	case qotcommon.KLType_KLType_1Min:
		return qotcommon.SubType_SubType_KL_1Min
	case qotcommon.KLType_KLType_Day:
		return qotcommon.SubType_SubType_KL_Day
	case qotcommon.KLType_KLType_Week:
		return qotcommon.SubType_SubType_KL_Week
	case qotcommon.KLType_KLType_Month:
		return qotcommon.SubType_SubType_KL_Month
	case qotcommon.KLType_KLType_Year:
		return qotcommon.SubType_SubType_KL_Year
	case qotcommon.KLType_KLType_Quarter:
		return qotcommon.SubType_SubType_KL_Qurater
	case qotcommon.KLType_KLType_5Min:
		return qotcommon.SubType_SubType_KL_5Min
	case qotcommon.KLType_KLType_15Min:
		return qotcommon.SubType_SubType_KL_15Min
	case qotcommon.KLType_KLType_30Min:
		return qotcommon.SubType_SubType_KL_30Min
	case qotcommon.KLType_KLType_60Min:
		return qotcommon.SubType_SubType_KL_60Min
	case qotcommon.KLType_KLType_3Min:
		return qotcommon.SubType_SubType_KL_3Min
	default:
		return qotcommon.SubType_SubType_None
	}
}
func (f *FTApi) GetHistoryKline(ctx context.Context, sec *qotcommon.Security, klType qotcommon.KLType, begin string, nextKey []byte) (*qotrequesthistorykl.S2C, error) {
	//sub first
	subType := klTypeToSubType(klType)
	if subType == qotcommon.SubType_SubType_None {
		return nil, fmt.Errorf("invalid klType: %d", klType)
	}
	_, err := f.SubSec(ctx, sec, subType)
	if err != nil {
		return nil, err
	}
	var end = time.Now().Format("2006-01-02")
	return f.fapi.RequestHistoryKLine(ctx,
		sec, begin, end, klType,
		qotcommon.RehabType_RehabType_Forward,
		&futuapi.OptionalInt32{Value: math.MaxInt32},
		qotcommon.KLFields_KLFields_None, nextKey,
		&futuapi.OptionalBool{Value: false},
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
func (f *FTApi) PlaceOrder(ctx context.Context, acc *trdcommon.TrdAcc, trdSide trdcommon.TrdSide, sec *qotcommon.Security, qty float64, price float64, remark string, orderType trdcommon.OrderType) (*trdplaceorder.S2C, error) {
	return f.fapi.PlaceOrder(ctx,
		&trdcommon.TrdHeader{
			TrdEnv:    acc.TrdEnv,
			TrdMarket: sec.Market,
			AccID:     acc.AccID,
		},
		trdSide, orderType, *sec.Code, qty,
		&futuapi.OptionalDouble{Value: price},
		&futuapi.OptionalBool{Value: true},
		&futuapi.OptionalDouble{Value: 0},
		trdcommon.TrdSecMarket(*sec.Market), remark,
		trdcommon.TimeInForce_TimeInForce_GTC,
		&futuapi.OptionalBool{Value: true},
		&futuapi.OptionalDouble{Value: price},
		trdcommon.TrailType_TrailType_Ratio,
		&futuapi.OptionalDouble{Value: 1},
		&futuapi.OptionalDouble{Value: 0},
	)
}
func (f *FTApi) PlaceBuyOrder(ctx context.Context, acc *trdcommon.TrdAcc, sec *qotcommon.Security, qty float64, price float64, remark string) (*trdplaceorder.S2C, error) {
	return f.fapi.PlaceOrder(ctx,
		&trdcommon.TrdHeader{
			TrdEnv:    acc.TrdEnv,
			TrdMarket: sec.Market,
			AccID:     acc.AccID,
		},
		trdcommon.TrdSide_TrdSide_Buy,
		trdcommon.OrderType_OrderType_Normal, *sec.Code, qty,
		&futuapi.OptionalDouble{Value: price},
		&futuapi.OptionalBool{Value: true},
		&futuapi.OptionalDouble{Value: 0},
		trdcommon.TrdSecMarket(*sec.Market), remark,
		trdcommon.TimeInForce_TimeInForce_GTC,
		&futuapi.OptionalBool{Value: true},
		&futuapi.OptionalDouble{Value: price},
		trdcommon.TrailType_TrailType_Ratio,
		&futuapi.OptionalDouble{Value: 1},
		&futuapi.OptionalDouble{Value: 0},
	)
}
func (f *FTApi) PlaceSellOrder(ctx context.Context, acc *trdcommon.TrdAcc, sec *qotcommon.Security, qty float64, price float64, remark string) (*trdplaceorder.S2C, error) {
	return f.fapi.PlaceOrder(ctx,
		&trdcommon.TrdHeader{
			TrdEnv:    acc.TrdEnv,
			TrdMarket: sec.Market,
			AccID:     acc.AccID,
		},
		trdcommon.TrdSide_TrdSide_Sell,
		trdcommon.OrderType_OrderType_Normal, *sec.Code, qty,
		&futuapi.OptionalDouble{Value: price},
		&futuapi.OptionalBool{Value: true},
		&futuapi.OptionalDouble{Value: 0},
		trdcommon.TrdSecMarket(*sec.Market), remark,
		trdcommon.TimeInForce_TimeInForce_GTC,
		&futuapi.OptionalBool{Value: true},
		&futuapi.OptionalDouble{Value: price},
		trdcommon.TrailType_TrailType_Ratio,
		&futuapi.OptionalDouble{Value: 1},
		&futuapi.OptionalDouble{Value: 0},
	)
}
func (f *FTApi) PlaceMktBuyOrder(ctx context.Context, acc *trdcommon.TrdAcc, sec *qotcommon.Security, qty float64, price float64, remark string) (*trdplaceorder.S2C, error) {
	return f.fapi.PlaceOrder(ctx,
		&trdcommon.TrdHeader{
			TrdEnv:    acc.TrdEnv,
			TrdMarket: sec.Market,
			AccID:     acc.AccID,
		},
		trdcommon.TrdSide_TrdSide_Buy,
		trdcommon.OrderType_OrderType_Market, *sec.Code, qty,
		&futuapi.OptionalDouble{Value: price},
		&futuapi.OptionalBool{Value: true},
		&futuapi.OptionalDouble{Value: 0},
		trdcommon.TrdSecMarket(*sec.Market), remark,
		trdcommon.TimeInForce_TimeInForce_GTC,
		&futuapi.OptionalBool{Value: true},
		&futuapi.OptionalDouble{Value: price},
		trdcommon.TrailType_TrailType_Ratio,
		&futuapi.OptionalDouble{Value: 1},
		&futuapi.OptionalDouble{Value: 0},
	)
}
func (f *FTApi) PlaceMktSellOrder(ctx context.Context, acc *trdcommon.TrdAcc, sec *qotcommon.Security, qty float64, price float64, remark string) (*trdplaceorder.S2C, error) {
	return f.fapi.PlaceOrder(ctx,
		&trdcommon.TrdHeader{
			TrdEnv:    acc.TrdEnv,
			TrdMarket: sec.Market,
			AccID:     acc.AccID,
		},
		trdcommon.TrdSide_TrdSide_Sell,
		trdcommon.OrderType_OrderType_Market, *sec.Code, qty,
		&futuapi.OptionalDouble{Value: price},
		&futuapi.OptionalBool{Value: true},
		&futuapi.OptionalDouble{Value: 0},
		trdcommon.TrdSecMarket(*sec.Market), remark,
		trdcommon.TimeInForce_TimeInForce_GTC,
		&futuapi.OptionalBool{Value: true},
		&futuapi.OptionalDouble{Value: price},
		trdcommon.TrailType_TrailType_Ratio,
		&futuapi.OptionalDouble{Value: 1},
		&futuapi.OptionalDouble{Value: 0},
	)
}
func (f *FTApi) AddPriceReminder(ctx context.Context, sec *qotcommon.Security, price float64, note string) (int64, error) {
	return f.fapi.SetPriceReminder(ctx, sec,
		qotsetpricereminder.SetPriceReminderOp_SetPriceReminderOp_Add,
		0,
		qotcommon.PriceReminderType_PriceReminderType_PriceUp,
		qotcommon.PriceReminderFreq_PriceReminderFreq_OnceADay,
		&futuapi.OptionalDouble{Value: price}, note)
}
func (f *FTApi) SetPriceReminder(ctx context.Context, sec *qotcommon.Security, price float64, note string, key int64) (int64, error) {
	return f.fapi.SetPriceReminder(ctx, sec,
		qotsetpricereminder.SetPriceReminderOp_SetPriceReminderOp_Modify,
		key,
		qotcommon.PriceReminderType_PriceReminderType_PriceUp,
		qotcommon.PriceReminderFreq_PriceReminderFreq_OnceADay,
		&futuapi.OptionalDouble{Value: price}, note)
}
func (f *FTApi) DelPriceReminder(ctx context.Context, sec *qotcommon.Security, key int64) (int64, error) {
	return f.fapi.SetPriceReminder(ctx, sec,
		qotsetpricereminder.SetPriceReminderOp_SetPriceReminderOp_Del,
		key, 0, 0, nil, "")
}
func (f *FTApi) DelAllPriceReminder(ctx context.Context, sec *qotcommon.Security) (int64, error) {
	return f.fapi.SetPriceReminder(ctx, sec,
		qotsetpricereminder.SetPriceReminderOp_SetPriceReminderOp_DelAll,
		0, 0, 0, nil, "")
}
func (f *FTApi) SysNotify(ctx context.Context, callback func(interface{})) error {
	ch, err := f.fapi.SysNotify(ctx)
	if err != nil {
		return err
	}
	selectChan(ctx, ch, callback)
	return nil
}
func (f *FTApi) UpdateKL(ctx context.Context, callback func(interface{})) error {
	ch, err := f.fapi.UpdateKL(ctx)
	if err != nil {
		return err
	}
	selectChan(ctx, ch, callback)
	return nil
}
func (f *FTApi) UpdateDeal(ctx context.Context, callback func(interface{})) error {
	ch, err := f.fapi.UpdateDeal(ctx)
	if err != nil {
		return err
	}
	selectChan(ctx, ch, callback)
	return nil
}
func (f *FTApi) UpdateRT(ctx context.Context, callback func(interface{})) error {
	ch, err := f.fapi.UpdateRT(ctx)
	if err != nil {
		return err
	}
	selectChan(ctx, ch, callback)
	return nil
}
func (f *FTApi) UpdateBasicQot(ctx context.Context, callback func(interface{})) error {
	ch, err := f.fapi.UpdateBasicQot(ctx)
	if err != nil {
		return err
	}
	selectChan(ctx, ch, callback)
	return nil
}
func (f *FTApi) UpdateOrder(ctx context.Context, callback func(interface{})) error {
	ch, err := f.fapi.UpdateOrder(ctx)
	if err != nil {
		return err
	}
	selectChan(ctx, ch, callback)
	return nil
}
func (f *FTApi) UpdatePriceReminder(ctx context.Context, callback func(interface{})) error {
	ch, err := f.fapi.UpdatePriceReminder(ctx)
	if err != nil {
		return err
	}
	selectChan(ctx, ch, callback)
	return nil
}
func (f *FTApi) UpdateTicker(ctx context.Context, sec *qotcommon.Security, callback func(interface{})) error {
	unsub, err := f.SubSec(ctx, sec, qotcommon.SubType_SubType_Ticker)
	if err != nil {
		return err
	}
	defer unsub()
	ch, err := f.fapi.UpdateTicker(ctx)
	if err != nil {
		return err
	}
	selectChan(ctx, ch, callback)
	return nil
}
func (f *FTApi) TrdMarketName(i int32) string {
	return trdcommon.TrdMarket_name[i]
}
func (f *FTApi) TrdMarketNames(i ...int32) []string {
	names := make([]string, 0, len(i))
	for _, v := range i {
		names = append(names, trdcommon.TrdMarket_name[v])
	}
	return names
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
func (f *FTApi) TrdSecMarketNames(i ...int32) []string {
	names := make([]string, 0, len(i))
	for _, v := range i {
		names = append(names, trdcommon.TrdSecMarket_name[v])
	}
	return names
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
func (f *FTApi) MktCurrency(i int32) int32 {
	return int32(getMktCurrency(i))
}
func (f *FTApi) MktCurrencyName(i int32) string {
	return trdcommon.Currency_name[int32(getMktCurrency(i))]
}
func (f *FTApi) MktCurrencyNames(i ...int32) []string {
	names := make([]string, 0, len(i))
	for _, v := range i {
		names = append(names, trdcommon.Currency_name[int32(getMktCurrency(v))])
	}
	return names
}
func (f *FTApi) GetAccCurrencyNames(acc *trdcommon.TrdAcc) []string {
	return f.MktCurrencyNames(acc.TrdMarketAuthList...)
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
func (f *FTApi) MktStateName(i int32) string {
	return qotcommon.QotMarketState_name[i]
}
func (f *FTApi) PriceReminderTypeName(i int32) string {
	return qotcommon.PriceReminderType_name[i]
}
func (f *FTApi) ProgramStatusName(i int32) string {
	return common.ProgramStatusType_name[i]
}
func (f *FTApi) ModifyOrderOpName(i int32) string {
	return trdcommon.ModifyOrderOp_name[i]
}
func (f *FTApi) ExchTypeName(i int32) string {
	return qotcommon.ExchType_name[i]
}
func (f *FTApi) SecTypeName(i int32) string {
	return qotcommon.SecurityType_name[i]
}
func (f *FTApi) SecMarketName(i int32) string {
	return qotcommon.QotMarket_name[i]
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
