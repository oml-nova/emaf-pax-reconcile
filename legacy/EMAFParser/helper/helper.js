import { batchSettlementTypeMapper } from "../parser/batchSettlementTypeMapper.js"
import { recordTypeMapper } from "../parser/fileRecordTypeMapper.js"

function getRecordIdentifier(line) {
  return recordTypeMapper(line.slice(9, 12))
}

function getRecordNumber(line) {
  return line.slice(0, 9).trim()
}

function getBatchSettlementType(line) {
  return batchSettlementTypeMapper(line.slice(15, 17))
}

function convertAmount(transcationAmount) {
  const tranactionAmountArr = transcationAmount.split('');
  tranactionAmountArr.splice(9, 0, '.');
  return Number(tranactionAmountArr.join('')).toFixed(2)
}

export {
  getRecordIdentifier,
  getRecordNumber,
  getBatchSettlementType,
  convertAmount,
}