package pkg

import (
	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/weeaa/jito-go/pb"
)

// ConvertTransactionToProtobufPacket converts a solana-go Transaction to a pb.Packet.
func ConvertTransactionToProtobufPacket(transaction *solana.Transaction) (jito_pb.Packet, error) {
	data, err := transaction.MarshalBinary()
	if err != nil {
		return jito_pb.Packet{}, err
	}

	return jito_pb.Packet{
		Data: data,
		Meta: &jito_pb.Meta{
			Size:        uint64(len(data)),
			Addr:        "",
			Port:        0,
			Flags:       nil,
			SenderStake: 0,
		},
	}, nil
}

// ConvertBatchTransactionToProtobufPacket converts a slice of solana-go Transaction to a slice of pb.Packet.
func ConvertBatchTransactionToProtobufPacket(transactions []*solana.Transaction) ([]*jito_pb.Packet, error) {
	packets := make([]*jito_pb.Packet, 0, len(transactions))
	for _, tx := range transactions {
		packet, err := ConvertTransactionToProtobufPacket(tx)
		if err != nil {
			return nil, err
		}

		packets = append(packets, &packet)
	}

	return packets, nil
}

// ConvertProtobufPacketToTransaction converts a pb.Packet to a solana-go Transaction.
func ConvertProtobufPacketToTransaction(packet *jito_pb.Packet) (*solana.Transaction, error) {
	tx := &solana.Transaction{}
	err := tx.UnmarshalWithDecoder(bin.NewBorshDecoder(packet.Data))
	if err != nil {
		return nil, err
	}

	return tx, nil
}

// ConvertBatchProtobufPacketToTransaction converts a slice of pb.Packet to a slice of solana-go Transaction.
func ConvertBatchProtobufPacketToTransaction(packets []*jito_pb.Packet) ([]*solana.Transaction, error) {
	txs := make([]*solana.Transaction, 0, len(packets))
	for _, packet := range packets {
		tx, err := ConvertProtobufPacketToTransaction(packet)
		if err != nil {
			return nil, err
		}

		txs = append(txs, tx)
	}

	return txs, nil
}
