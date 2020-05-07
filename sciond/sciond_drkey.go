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

package sciond

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/sciond"

	"github.com/JordiSubira/drkeymockup/drkey"
	"github.com/JordiSubira/drkeymockup/drkey/protocol"
)

// Service mocks up drkey feature in sciond.Service
type Service struct {
	sciond.Service
}

// NewService returns a SCIOND API connection factory.
func NewService(name string) Service {
	return Service{sciond.NewService(name)}
}

// Connect returns sciond.Connector interface implementation which also exports a mock for DRKeyGetLvl2Key
func (s Service) Connect(ctx context.Context) (Connector, error) {
	conn, err := s.Service.Connect(ctx)
	return Connector{conn}, err
}

// Connector mocks up drkey feature in sciond.Connector
type Connector struct {
	sciond.Connector
}

// DRKeyGetLvl2Key mocks retrieving Lvl2Key operation for SCIOND API
func (c Connector) DRKeyGetLvl2Key(_ context.Context, meta drkey.Lvl2Meta, valTime uint32) (drkey.Lvl2Key, error) {

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
	duration := int64(time.Hour / time.Second)
	epoch := drkey.NewEpoch(valTime, valTime+uint32(duration))

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

// GetDefaultSCIONDAddress exports sciond.GetDefaultSCIONDAddress
func GetDefaultSCIONDAddress(ia *addr.IA) string {
	return sciond.GetDefaultSCIONDAddress(ia)
}

// Send exports sciond.Send
func Send(pld *sciond.Pld, conn net.Conn) error {
	return sciond.Send(pld, conn)
}
