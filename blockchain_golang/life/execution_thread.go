package life

import (
	"github.com/VladChernenko/UndchainCore/block"
	"github.com/VladChernenko/UndchainCore/globals"
	"github.com/VladChernenko/UndchainCore/structures"
	"github.com/VladChernenko/UndchainCore/utils"
)

func ExecutionThread() {

	globals.EXECUTION_THREAD_METADATA_HANDLER.RWMutex.RLock()

	epochHandler := globals.EXECUTION_THREAD_METADATA_HANDLER.Handler

	currentEpochIsFresh := utils.EpochStillFresh(&epochHandler)

	// Struct is {currentLeader,currentToVerify,infoAboutFinalBlocksInThisEpoch:{poolPubKey:{index,hash}}}
	//alignmentData := globals.EXECUTION_THREAD_METADATA_HANDLER.Handler.CurrentEpochAlignmentData

	shouldMoveToNextEpoch := false

	if epochHandler.LegacyEpochAlignmentData.Activated {
		// Stub
	} else if currentEpochIsFresh && epochHandler.CurrentEpochAlignmentData.Activated {
		// Stub
	}

	if !currentEpochIsFresh && !epochHandler.LegacyEpochAlignmentData.Activated {

		TryToFinishCurrentEpoch(&epochHandler.EpochDataHandler)

	}

	if shouldMoveToNextEpoch {

		SetupNextEpoch(&epochHandler.EpochDataHandler)

	}

}

func ExecuteBlock(block *block.Block) {

	if globals.EXECUTION_THREAD_METADATA_HANDLER.Handler.ExecutionData[block.Creator].Hash == block.PrevHash {
		// Stub
	}

}

func ExecuteTransaction(tx *structures.Transaction) {}

func TryToFinishCurrentEpoch(epochHandler *structures.EpochDataHandler) {}

func SetupNextEpoch(epochHandler *structures.EpochDataHandler) {}
