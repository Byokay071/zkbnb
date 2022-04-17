/*
 * Copyright © 2021 Zecrey Protocol
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package l1BlockMonitor

import (
	"errors"
	"fmt"
	"github.com/zecrey-labs/zecrey-legend/common/model/l2BlockEventMonitor"
	"github.com/zecrey-labs/zecrey-legend/common/model/l2TxEventMonitor"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"gorm.io/gorm"
)

type (
	L1BlockMonitorModel interface {
		CreateL1BlockMonitorTable() error
		DropL1BlockMonitorTable() error
		CreateL1BlockMonitor(tx *L1BlockMonitor) (bool, error)
		CreateL1BlockMonitorsInBatches(blockInfos []*L1BlockMonitor) (rowsAffected int64, err error)
		CreateMonitorsInfo(blockInfo *L1BlockMonitor, txEventMonitors []*l2TxEventMonitor.L2TxEventMonitor, blockEventMonitors []*l2BlockEventMonitor.L2BlockEventMonitor) (err error)
		GetL1BlockMonitors() (blockInfos []*L1BlockMonitor, err error)
		GetLatestL1BlockMonitor() (blockInfo *L1BlockMonitor, err error)
		GetL1BlockMonitorsByChainId(chainId uint8) (blockInfos []*L1BlockMonitor, err error)
		GetLatestHandledBlockByChainId(chainId uint8) (blockInfo *L1BlockMonitor, err error)
		GetL1BlockMonitorsByChainIdAndL1BlockHeight(chainId uint8, l1BlockHeight int64) (blockInfo *L1BlockMonitor, err error)
	}

	defaultL1BlockMonitorModel struct {
		sqlc.CachedConn
		table string
		DB    *gorm.DB
	}

	L1BlockMonitor struct {
		gorm.Model
		// chain id
		ChainId uint8
		// l1 block height
		L1BlockHeight int64
		// block info, array of hashes
		BlockInfo string
	}
)

func (*L1BlockMonitor) TableName() string {
	return TableName
}

func NewL1BlockMonitorModel(conn sqlx.SqlConn, c cache.CacheConf, db *gorm.DB) L1BlockMonitorModel {
	return &defaultL1BlockMonitorModel{
		CachedConn: sqlc.NewConn(conn, c),
		table:      TableName,
		DB:         db,
	}
}

/*
	Func: CreateL1BlockMonitorTable
	Params:
	Return: err error
	Description: create l2 tx event monitor table
*/
func (m *defaultL1BlockMonitorModel) CreateL1BlockMonitorTable() error {
	return m.DB.AutoMigrate(L1BlockMonitor{})
}

/*
	Func: DropL1BlockMonitorTable
	Params:
	Return: err error
	Description: drop l2 tx event monitor table
*/
func (m *defaultL1BlockMonitorModel) DropL1BlockMonitorTable() error {
	return m.DB.Migrator().DropTable(m.table)
}

/*
	Func: CreateL1BlockMonitor
	Params: asset *L1BlockMonitor
	Return: bool, error
	Description: create L1BlockMonitor tx
*/
func (m *defaultL1BlockMonitorModel) CreateL1BlockMonitor(tx *L1BlockMonitor) (bool, error) {
	dbTx := m.DB.Table(m.table).Create(tx)
	if dbTx.Error != nil {
		err := fmt.Sprintf("[l1BlockMonitor.CreateL1BlockMonitor] %s", dbTx.Error)
		logx.Error(err)
		return false, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		ErrInvalidL1BlockMonitor := errors.New("invalid l1BlockMonitor")
		err := fmt.Sprintf("[l1BlockMonitor.CreateL1BlockMonitor] %s", ErrInvalidL1BlockMonitor)
		logx.Error(err)
		return false, ErrInvalidL1BlockMonitor
	}
	return true, nil
}

/*
	Func: CreateL1BlockMonitorsInBatches
	Params: []*L1BlockMonitor
	Return: rowsAffected int64, err error
	Description: create L1BlockMonitor batches
*/
func (m *defaultL1BlockMonitorModel) CreateL1BlockMonitorsInBatches(blockInfos []*L1BlockMonitor) (rowsAffected int64, err error) {
	dbTx := m.DB.Table(m.table).CreateInBatches(blockInfos, len(blockInfos))
	if dbTx.Error != nil {
		err := fmt.Sprintf("[l1BlockMonitor.CreateL1AssetsMonitorInBatches] %s", dbTx.Error)
		logx.Error(err)
		return 0, dbTx.Error
	}
	if dbTx.RowsAffected == 0 {
		return 0, nil
	}
	return dbTx.RowsAffected, nil
}

func (m *defaultL1BlockMonitorModel) CreateMonitorsInfo(
	blockInfo *L1BlockMonitor,
	txEventMonitors []*l2TxEventMonitor.L2TxEventMonitor,
	blockEventMonitors []*l2BlockEventMonitor.L2BlockEventMonitor) (err error) {
	err = m.DB.Transaction(
		func(tx *gorm.DB) error { // transact
			// create data for l1 block info
			dbTx := tx.Table(m.table).Create(blockInfo)
			if dbTx.Error != nil {
				return dbTx.Error
			}
			if dbTx.RowsAffected == 0 {
				return errors.New("[CreateMonitorsInfo] unable to create l1 block info")
			}
			// create data in batches for l2 tx event monitor
			dbTx = tx.Table(l2TxEventMonitor.TableName).CreateInBatches(txEventMonitors, len(txEventMonitors))
			if dbTx.Error != nil {
				return dbTx.Error
			}
			if dbTx.RowsAffected != int64(len(txEventMonitors)) {
				return errors.New("[CreateMonitorsInfo] unable to create l2 tx event monitors")
			}
			// create data in batches for l2 block event monitor
			dbTx = tx.Table(l2BlockEventMonitor.TableName).CreateInBatches(blockEventMonitors, len(blockEventMonitors))
			if dbTx.Error != nil {
				return dbTx.Error
			}
			if dbTx.RowsAffected != int64(len(blockEventMonitors)) {
				return errors.New("[CreateMonitorsInfo] unable to create l2 block event monitors")
			}
			return nil
		},
	)
	return err
}

/*
	GetL1BlockMonitors: get all L1BlockMonitors
*/
func (m *defaultL1BlockMonitorModel) GetL1BlockMonitors() (blockInfos []*L1BlockMonitor, err error) {
	dbTx := m.DB.Table(m.table).Find(&blockInfos).Order("l1_block_height")
	if dbTx.Error != nil {
		err := fmt.Sprintf("[l1BlockMonitor.GetL1BlockMonitors] %s", dbTx.Error)
		logx.Error(err)
		return nil, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		err := fmt.Sprintf("[l1BlockMonitor.GetL1BlockMonitors] %s", ErrNotFound)
		logx.Error(err)
		return nil, ErrNotFound
	}
	return blockInfos, dbTx.Error
}

/*
	Func: GetLatestL1BlockMonitor
	Return: blockInfos []*L1BlockMonitor, err error
	Description: get latest l1 block monitor info
*/
func (m *defaultL1BlockMonitorModel) GetLatestL1BlockMonitor() (blockInfo *L1BlockMonitor, err error) {
	dbTx := m.DB.Table(m.table).Order("l1_block_height desc").First(&blockInfo)
	if dbTx.Error != nil {
		err := fmt.Sprintf("[l1BlockMonitor.GetLatestL1BlockMonitor] %s", dbTx.Error)
		logx.Error(err)
		return nil, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		err := fmt.Sprintf("[l1BlockMonitor.GetLatestL1BlockMonitor] %s", ErrNotFound)
		logx.Error(err)
		return nil, ErrNotFound
	}
	return blockInfo, nil
}

/*
	Func: GetL1BlockMonitorsByChainId
	Return: blockInfos []*L1BlockMonitor, err error
	Description: get L1BlockMonitor by chain id
*/
func (m *defaultL1BlockMonitorModel) GetL1BlockMonitorsByChainId(chainId uint8) (blockInfos []*L1BlockMonitor, err error) {
	dbTx := m.DB.Table(m.table).Where("chain_id = ?", chainId).Find(&blockInfos).Order("l1_block_height desc")
	if dbTx.Error != nil {
		err := fmt.Sprintf("[l1BlockMonitor.GetL1BlockMonitorsByChainId] %s", dbTx.Error)
		logx.Error(err)
		return nil, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		err := fmt.Sprintf("[l1BlockMonitor.GetL1BlockMonitorsByChainId] %s", ErrNotFound)
		logx.Error(err)
		return nil, ErrNotFound
	}
	return blockInfos, nil
}

/*
	Func: GetLatestHandledBlockByChainId
	Return: blockInfos *L1BlockMonitor, err error
	Description: get latest handled block by chain id
*/
func (m *defaultL1BlockMonitorModel) GetLatestHandledBlockByChainId(chainId uint8) (blockInfo *L1BlockMonitor, err error) {
	dbTx := m.DB.Table(m.table).Where("chain_id = ?", chainId).Order("l1_block_height desc").Find(&blockInfo)
	if dbTx.Error != nil {
		err := fmt.Sprintf("[l1BlockMonitor.GetL1BlockMonitorsByChainId] %s", dbTx.Error)
		logx.Error(err)
		return nil, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		err := fmt.Sprintf("[l1BlockMonitor.GetL1BlockMonitorsByChainId] %s", ErrNotFound)
		logx.Error(err)
		return nil, ErrNotFound
	}
	return blockInfo, nil
}

/*
	Func: GetL1BlockMonitorsByChainIdAndL1BlockHeight
	Return: blockInfos []*L1BlockMonitor, err error
	Description: get L1BlockMonitor by chain id and l1 block height
*/
func (m *defaultL1BlockMonitorModel) GetL1BlockMonitorsByChainIdAndL1BlockHeight(chainId uint8, l1BlockHeight int64) (blockInfo *L1BlockMonitor, err error) {
	dbTx := m.DB.Table(m.table).Where("chain_id = ? AND l1_block_height = ?", chainId, l1BlockHeight).First(&blockInfo)
	if dbTx.Error != nil {
		err := fmt.Sprintf("[l1BlockMonitor.GetL1BlockMonitorsByChainIdAndL1BlockHeight] %s", dbTx.Error)
		logx.Error(err)
		return nil, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		err := fmt.Sprintf("[l1BlockMonitor.GetL1BlockMonitorsByChainIdAndL1BlockHeight] %s", ErrNotFound)
		logx.Error(err)
		return nil, ErrNotFound
	}
	return blockInfo, nil
}
