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

package nft

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"gorm.io/gorm"
)

type (
	AccountL2NftModel interface {
		CreateAccountL2NftTable() error
		DropAccountL2NftTable() error
	}
	defaultAccountL2NftModel struct {
		sqlc.CachedConn
		table string
		DB    *gorm.DB
	}

	AccountL2Nft struct {
		gorm.Model
		NftIndex            int64 `gorm:"uniqueIndex"`
		CreatorAccountIndex int64
		OwnerAccountIndex   int64
		NftAccountIndex     int64
		AssetId             int64
		AssetAmount         int64
		NftContentHash      string
		NftUrl              string
		ChainId             int64
		NftL1TokenId        int64
		NftL1Address        string
		NftDetail           *AccountL2NftDetail `gorm:"foreignkey:NftId"`
		CollectionId        int64
		Status              int
	}
)

func NewAccountL2NftModel(conn sqlx.SqlConn, c cache.CacheConf, db *gorm.DB) AccountL2NftModel {
	return &defaultAccountL2NftModel{
		CachedConn: sqlc.NewConn(conn, c),
		table:      AccountL2NftTableName,
		DB:         db,
	}
}

func (*AccountL2Nft) TableName() string {
	return AccountL2NftTableName
}

/*
	Func: CreateAccountL2NftTable
	Params:
	Return: err error
	Description: create account l2 nft table
*/
func (m *defaultAccountL2NftModel) CreateAccountL2NftTable() error {
	return m.DB.AutoMigrate(AccountL2Nft{})
}

/*
	Func: DropAccountNFTTable
	Params:
	Return: err error
	Description: drop accountnft table
*/
func (m *defaultAccountL2NftModel) DropAccountL2NftTable() error {
	return m.DB.Migrator().DropTable(m.table)
}
