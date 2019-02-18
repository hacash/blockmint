package block1def

/**
 * block format field byte size define
 *
 */

var (
	ByteSizeBlockHead              = 1 + 5 + 5 + 32 + 32 + 4
	ByteSizeBlockMeta              = 4
	ByteSizeBlockBeforeTransaction = ByteSizeBlockHead + ByteSizeBlockMeta

	// block base

	ByteSize_version   = 1 // 0 ~ 255
	ByteSize_height    = 5 // 0 ~ 255 * 4294967295
	ByteSize_timestamp = 5

	ByteSize_prevHash = 32 // hash256
	ByteSize_mrklRoot = 32 // hash256

	ByteSize_extensionSize      = 2 // 0 ~ 65535
	ByteSizeWidth_extensionSize = 4 // 1 size = 4 byte

	// transaction

	ByteSize_transactionHash = 32 // hash256

	ByteSize_transactionCount     = 4 // 0 ~ 4294967295
	ByteSize_trs_type             = 1 // 0 ~ 255
	ByteSize_trs_address          = 34
	ByteSize_trs_timestamp        = 5
	ByteSize_trs_coinbase_message = 16

	ByteSize_trs_billCount   = 1 // 0 ~ 255
	ByteSize_trs_bill_unit   = 1 // 0 ~ 255
	ByteSize_trs_bill_amount = 3 // 0 ~ 16777216

	ByteSize_trs_fee_unit   = 1 // 0 ~ 255
	ByteSize_trs_fee_amount = 3 // 0 ~ 16777216

	ByteSize_trs_appendixSize      = 1 // 0 ~ 255
	ByteSizeWidth_trs_appendixSize = 4 // 1 size = 4 byte

	ByteSize_actionCount = 1 // 0 ~ 255

	// action

	ByteSize_action_kind       = 2 // 0 ~ 65535
	ByteSize_inputAddressCount = 1 // 0 ~ 255
	ByteSize_outputCount       = 1 // 0 ~ 255

	// liquidation

	ByteSize_parteEndLockHeightNum    = 2 // 0 ~ 65535
	ByteSize_liquidationId            = 8 //
	ByteSize_liquidationCount         = 1
	ByteSize_liquidationAutoincrement = 4 //

	// group address

	ByteSize_validRightsRatio = 2 //
	ByteSize_group_formCount  = 1 // 0 ~ 200
	ByteSize_group_rights     = 4 // vote etc.

	// diamond
	ByteSize_diamond_face  = 6 // WWAASS
	ByteSize_diamond_nonce = 8 //

	// sign

	ByteSize_signCount = 2 // 0 ~ 65535
	ByteSize_publicKey = 33
	ByteSize_signature = 64

	ByteSize_multisignCount        = 2 // 0 ~ 65535
	ByteSizeLength_publicKeyScript = 1 // maxnum = 200
	ByteSizeLength_signatureScript = 1 // maxnum = 200

)
