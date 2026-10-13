package outbound

import (
	"context"
	"math/big"
)

// Follow the serial captured by the originating pick/pack movement, even when
// several serials now share one pooled staging or packing balance.
func (s *Service) moveDocumentSerial(ctx context.Context, source string, target *string, qty *big.Rat, serialID *string) (*string, error) {
	if serialID == nil {
		return s.moveSerialIdentity(ctx, source, target, qty)
	}
	if qty.Cmp(big.NewRat(1, 1)) != 0 {
		return nil, state("a serialized document line must move exactly one base unit")
	}
	current, err := s.repositories.Inventory.SerialState.GetLocked(ctx, *serialID)
	if err != nil {
		return nil, err
	}
	if current.BalanceID != source {
		return nil, state("document serial no longer belongs to its source balance")
	}
	if target == nil {
		err = s.repositories.Inventory.SerialState.Delete(ctx, *serialID)
	} else {
		err = s.repositories.Inventory.SerialState.Move(ctx, *serialID, *target)
	}
	if err != nil {
		return nil, err
	}
	return serialID, nil
}
