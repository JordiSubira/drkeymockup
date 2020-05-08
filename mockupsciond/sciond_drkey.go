// Copyright 2020 ETH Zurich
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package mockupsciond

import (
	"context"
	"fmt"
	"time"

	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/common"

	"github.com/JordiSubira/drkeymockup/drkey"
	"github.com/JordiSubira/drkeymockup/drkey/protocol"
)

const keyDuration = time.Hour * 24

// DRKeyGetLvl2Key mocks retrieving Lvl2Key operation for SCIOND API
func DRKeyGetLvl2Key(_ context.Context, meta drkey.Lvl2Meta, valTime uint32) (drkey.Lvl2Key, error) {
	lvl1Key, err := getLvl1(meta.SrcIA, meta.DstIA, valTime)
	if err != nil {
		return drkey.Lvl2Key{}, common.NewBasicError("Error getting lvl1 key", err)
	}

	derProt, found := protocol.KnownDerivations[meta.Protocol]
	if !found {
		return drkey.Lvl2Key{}, fmt.Errorf("No derivation found for protocol \"%s\"", meta.Protocol)
	}
	return derProt.DeriveLvl2(meta, lvl1Key)

}

func getLvl1(srcIA, dstIA addr.IA, valTime uint32) (drkey.Lvl1Key, error) {
	validSince := uint32(time.Unix(int64(valTime), 0).Truncate(keyDuration).Unix())
	duration := uint32(keyDuration / time.Second)
	epoch := drkey.NewEpoch(validSince, validSince+duration)

	meta := drkey.SVMeta{
		Epoch: epoch,
	}
	asSecret := []byte{0, 1, 2, 3, 4, 5, 6, 7, 0, 1, 2, 3, 4, 5, 6, 7}
	sv, err := drkey.DeriveSV(meta, asSecret)
	if err != nil {
		return drkey.Lvl1Key{}, common.NewBasicError("Error getting secret value", err)
	}

	lvl1, err := protocol.DeriveLvl1(drkey.Lvl1Meta{
		Epoch: epoch,
		SrcIA: srcIA,
		DstIA: dstIA,
	}, sv)
	if err != nil {
		return drkey.Lvl1Key{}, common.NewBasicError("Error deriving Lvl1 key", err)
	}

	return lvl1, nil
}
