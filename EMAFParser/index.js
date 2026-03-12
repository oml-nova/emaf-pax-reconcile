import readLine from 'readline';
import * as helper from './helper/helper.js';
import { CreditHeader } from './parser/creditHeader.js';
import { MerchantHeader1 } from './parser/merchantHeader1.js';
import { MerchantHeader2 } from './parser/merchantHeader2.js';
import { CreditReconciliation1 } from './parser/creditReconciliation1.js';
import { CreditReconciliation2 } from './parser/creditReconciliation2.js';
import { CreditReconciliation3 } from './parser/creditReconciliation3.js';
import { CreditReconciliation4 } from './parser/creditReconciliation4.js';
import { QueueService } from './service/SQSService.js'
import { StorageService } from './service/S3Service.js'
import { getCardNetwork, getCardNetworkType } from './parser/cardTypeMapper.js';
import { v4 as uuidv4 } from 'uuid';

function generateUUID() {
  return uuidv4();
}

function flushTransaction(transaction, merchantRecords, dataBatch, queuePromises) {
  try {
    const data = {
      merchantId: merchantRecords['settlement_mid'],
      transaction
    }
    if (dataBatch.length === 10) {
      queuePromises.push(QueueService.sendBatch(JSON.parse(JSON.stringify(dataBatch))))
      dataBatch.length = 0;
    }
    dataBatch.push({
      Id: generateUUID(),
      MessageBody: JSON.stringify(data)
    });
  } catch (error) {
    console.error("Error pushing to SQS", error)
  }
}

export const handler = async (event) => {
  const bucket = event.Records[0].s3.bucket.name;
  const key = decodeURIComponent(event.Records[0].s3.object.key.replace(/\+/g, ' '));
  try {
    const params = {
      Bucket: bucket,
      Key: key,
    };
    const { Body } = await StorageService.getFile(params);
    const readLineStream = readLine.createInterface({
      input: Body,
      crlfDelay: Infinity
    });

    let merchantRecords = {
      transactions: []
    };
    let transaction = {}
    let dataBatch = [];
    const queuePromises = [];

    for await (const line of readLineStream) {
      //File Header Records
      if (helper.getRecordIdentifier(line) === "CREDIT_FILE_HEADER") {
        const creditHeader = new CreditHeader(line);
        merchantRecords['record_type'] = "CREDIT";
        merchantRecords['version'] = creditHeader.getFileVersion();
        merchantRecords['chain_code'] = creditHeader.getChainCode();
        merchantRecords['file_creation_date'] = creditHeader.getCreationDateTime();
      }

      //Batch Header Record 1
      if (helper.getRecordIdentifier(line) === "MID_HEADER_1") {
        const merchantHeader1 = new MerchantHeader1(line);
        merchantRecords['batch_settlement_type'] = merchantHeader1.getBatchSettlementType();
        merchantRecords['settlement_mid'] = merchantHeader1.getSettlementMID();
        merchantRecords['frontend_mid'] = merchantHeader1.getFrontEndMID();
        merchantRecords['division_number'] = merchantHeader1.getMerchantDivisionNumber();
        merchantRecords['store_number'] = merchantHeader1.getMerchantStoreNumber();
        merchantRecords['merchant_name'] = merchantHeader1.getMerchantName();
        merchantRecords['merchant_countrycode'] = merchantHeader1.getMerchantLocationCountryCode();
        merchantRecords['batch_number_file_identifier'] = merchantHeader1.getBatchNumberFileIdentifier();
      }

      if (helper.getRecordIdentifier(line) === "MID_HEADER_2") {
        const merchantHeader2 = new MerchantHeader2(line);
        merchantRecords['merchant_name'] = merchantRecords['merchant_name'] ? merchantRecords['merchant_name'] : merchantHeader2.getMerchantName()
        merchantRecords['merchant_city'] = merchantHeader2.getMerchantCity();
        merchantRecords['merchant_state'] = merchantHeader2.getMerchantState();
        merchantRecords['merchant_zipcode'] = merchantHeader2.getMerchantZipCode();
        merchantRecords['merchant_countrycode'] = merchantRecords['merchant_countrycode'] ? merchantRecords['merchant_countrycode'] : merchantHeader2.getMerchantCountryCode();
      }

      if (helper.getRecordIdentifier(line) === "CREDIT_RECONCILIATION_1") {
        if (Object.keys(transaction).length) {
          flushTransaction(transaction, merchantRecords, dataBatch, queuePromises);
          merchantRecords.transactions.push(transaction);
          transaction = {}
        }
        const creditReconciliation1 = new CreditReconciliation1(line);
        transaction['date_time'] = creditReconciliation1.getTransactionDateTime();
        transaction['sequence_no'] = creditReconciliation1.getTransactionSequenceNumber();
        transaction['type_code'] = creditReconciliation1.getTransactionTypeCode();
        transaction['authorization_code'] = creditReconciliation1.getAuthorizationCode();
        transaction['entry_mode'] = creditReconciliation1.getPOSEntryMode();
        transaction['account_number'] = creditReconciliation1.getCardAccountNumber();
        transaction['expiry'] = creditReconciliation1.getExpirationDate();
        transaction['amount'] = creditReconciliation1.getTransactionAmount();
        transaction['old_authorized_amount'] = creditReconciliation1.getOldAuthorizationAmount();
        transaction['cashback_amount'] = creditReconciliation1.getCashBackAmount();
        transaction['mcc_sic_code'] = creditReconciliation1.getMCCSICCode();
        transaction['card_network_type'] = getCardNetwork(creditReconciliation1.getCardNetworkType());
        transaction['card_type'] = getCardNetworkType(creditReconciliation1.getCardNetworkType());
        transaction['product_type'] = creditReconciliation1.getCardProductType();
        transaction['draft_locator_no'] = creditReconciliation1.getDraftLocatorNumber();
        transaction['batch_number'] = creditReconciliation1.getBatchNumber();
        transaction['convenience_fee'] = creditReconciliation1.getConvenincetFee();
        transaction['network_reference_number'] = creditReconciliation1.getNetworkReferenceNumber();
        transaction['terminal_no'] = creditReconciliation1.getTerminalNumber();
        transaction['tic'] = creditReconciliation1.getTransactionIntegrityClassification();
        transaction['tid'] = creditReconciliation1.getTerminalIdentifier();
        transaction['pin_optimization'] = creditReconciliation1.getPinOptimisationFlag();
      }

      if (helper.getRecordIdentifier(line) === "CREDIT_RECONCILIATION_2") {
        const creditReconciliation2 = new CreditReconciliation2(line);
        transaction['authorization_source'] = creditReconciliation2.getAuthorizationSource();
        transaction['indicator'] = creditReconciliation2.getMailPhoneIndicator();
        transaction['cat'] = creditReconciliation2.getCATIndicator();
        transaction['avs_response'] = creditReconciliation2.getASVResponseCode();
        transaction['cvv2'] = creditReconciliation2.getCVV2ResponseIndicator();
        transaction['reg_no'] = creditReconciliation2.getRegistrationNumber();
        transaction['merchant_supplied_data'] = creditReconciliation2.getMerchatSuppliedData();
        transaction['currency'] = creditReconciliation2.getCurrencyCode();
        transaction['aci'] = creditReconciliation2.getAuthorizationCharacteristicIndicator();
        transaction['interchange_code'] = creditReconciliation2.getInterChangeCode();
        transaction['interchange_amount'] = creditReconciliation2.getInterChangeSign() + creditReconciliation2.getInterChangeReimbursenebtFee();
        transaction['surcharge_code'] = creditReconciliation2.getSurChargeReasonOrInterchangeAdjustmentReasonCode();
        transaction['surcharge_adjustment_amount'] = creditReconciliation2.getSurchargeInterChangeAdjustmentSign() + creditReconciliation2.getSurChargeOrInterchangeAdjustmentAmount();
        transaction['transaction_id'] = creditReconciliation2.getVisaTransactionId();
        transaction['validation_code'] = creditReconciliation2.getVisaValidationCode();
        transaction['auth_code'] = creditReconciliation2.getVisaAuthorizationCode();
        transaction['banknet_no'] = creditReconciliation2.getMasterCardBanknetRefNumber();
        transaction['banknet_settlement_date'] = creditReconciliation2.getMasterCardBanknetSettlementDate();
        transaction['visa_product_code'] = creditReconciliation2.getVisaProductCode();
        transaction['rtc_settlement_type'] = creditReconciliation2.getRTCSettlementType();
        transaction['token'] = creditReconciliation2.getToken();
        transaction['token_id'] = creditReconciliation2.getTokenId();
        transaction['token_expiry'] = creditReconciliation2.getTokenExpirationDate();
        transaction['token_assurance_level'] = creditReconciliation2.getTokenAssuranceLevel();
        transaction['account_ending_number'] = creditReconciliation2.getPANEndingNumber();
        transaction['evm_transaction_indicator'] = creditReconciliation2.getEVMTransactionIndicator();
        transaction['account_ref_no'] = creditReconciliation2.getPaymentAccountReferenceNumber();
      }

      if (helper.getRecordIdentifier(line) === "CREDIT_RECONCILIATION_3") {
        const creditReconciliation3 = new CreditReconciliation3(line);
        transaction['amount'] = transaction['amount'] ? transaction['amount'] : creditReconciliation3.getTransactionAmount();
        transaction['currency'] = transaction['currency'] ? transaction['currency'] : creditReconciliation3.getTransactionAmountCurrencyCode();
        transaction['dcc_mcp'] = creditReconciliation3.getDCCMCPIndicator();
        transaction['tip_gratuity_amount'] = creditReconciliation3.getTipORGratuityAmount();
      }

      if (helper.getRecordIdentifier(line) === "CREDIT_RECONCILIATION_4") {
        const creditReconciliation4 = new CreditReconciliation4(line);
        transaction['token'] = creditReconciliation4.getToken();
        transaction['token_id'] = creditReconciliation4.getTokenId();
        transaction['amount'] = transaction['amount'] ? transaction['amount'] : creditReconciliation4.getTransactionAmount();
        transaction['action_type'] = creditReconciliation4.getActionType()
      }

      //File Trailer Records
      if (helper.getRecordIdentifier(line) === "CREDIT_FILE_END") {
        // Flush the last transaction before resetting merchantRecords
        if (Object.keys(transaction).length) {
          flushTransaction(transaction, merchantRecords, dataBatch, queuePromises);
          merchantRecords.transactions.push(transaction);
          transaction = {}
        }
        merchantRecords = {
          transactions: []
        }
      }
    }

    // Flush the last transaction if file didn't end with CREDIT_FILE_END
    if (Object.keys(transaction).length) {
      flushTransaction(transaction, merchantRecords, dataBatch, queuePromises);
    }

    // Flush any remaining messages in the batch
    if (dataBatch.length > 0) {
      queuePromises.push(QueueService.sendBatch(JSON.parse(JSON.stringify(dataBatch))))
    }

    console.log("Parsing Completed For File", key);

    try {
      await Promise.allSettled(queuePromises);
      console.log("Data Pushed To Queue");
      return { statusCode: 200, body: 'File parsed successfully' };
    } catch (err) {
      console.error('Error:', err);
      return { statusCode: 500, body: 'Error uploading file' };
    }
  } catch (err) {
    console.log(`Error getting object ${key} from bucket ${bucket}. Make sure they exist and your bucket is in the same region as this function.`, err);
    return { statusCode: 500, body: `Error Processing File ${key} from ${bucket}` };
  }
};