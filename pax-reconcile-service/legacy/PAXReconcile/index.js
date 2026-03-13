import { MongoClient } from 'mongodb';
import { v4 as uuidv4 } from 'uuid';
import { createMechanism } from '@jm18457/kafkajs-msk-iam-authentication-mechanism';
import { Kafka } from 'kafkajs';


let merchantMapping = {};
let cachedDb;
let client;
let kafkaClient;
let producer;

export const connectToDatabase = async () => {
  if (client && client?.topology?.isConnected()) {
    console.log("Existing cached connection found!");
    return cachedDb;
  }
  console.log("Aquiring new DB connection....");
  try {
    client = await MongoClient.connect(process.env.DB_URI, {
      readPreference: 'secondaryPreferred'
    });
    const db = client.db(process.env.DB_NAME);

    cachedDb = db;
    return db;
  } catch (error) {
    console.log("ERROR aquiring DB Connection!");
    console.log(error);
    throw error;
  }
};

export const connectToKafka = async () => {
  const kafkaConfig = {
    clientId: 'REPORT_SERVICE_KAFKA',
    brokers: JSON.parse(`${process.env.KAFKA_HOST}`),
    requestTimeout: 450000,
    authenticationTimeout: 400000,
    reauthenticationThreshold: 40000,
    connectionTimeout: 3000,
  };

  if (process.env.KAFKA_SSL === 'true') {
    kafkaConfig.ssl = true;
    kafkaConfig.sasl = createMechanism({ region: 'ap-south-1' });
  }
  if (process.env.KAFKA_SASL === 'true') {
    kafkaConfig.sasl = {
      mechanism: 'plain',
      username: process.env.KAFKA_SASL_USERNAME,
      password: process.env.KAFKA_SASL_PASSWORD,
    };
  }
  if (!producer) {
    kafkaClient = new Kafka(kafkaConfig);
  }

  producer = kafkaClient.producer();
  await producer.connect();
}

const sendMessage = async (eventType, data) => {
  await producer.send({
    topic: process.env.PRODUCER_TOPIC,
    messages: [{ value: JSON.stringify({ type: eventType, data }) }]
  })
}

function generateUUID() {
  return uuidv4();
}

async function getRestaurantRefIdByPaxMerchantId(merchantId) {
  const connection = await connectToDatabase();
  if (merchantMapping[merchantId]?.fetched) {
    return merchantMapping[merchantId].refId
  }
  const merchantDetails = await connection.collection('merchant_accounts').findOne({ gatewayId: "WorldPay", merchantId });
  merchantMapping[merchantId] = { refId: merchantDetails?.refId ?? null, fetched: true }
  return merchantMapping[merchantId].refId
}

async function mapTransaction(data = { merchantId: "", transaction: {} }) {
  const connection = await connectToDatabase();
  const restaurantRefId = await getRestaurantRefIdByPaxMerchantId(data.merchantId);
  const amount = Number(data.transaction.amount ?? 0)
  const date = new Date(data.transaction.date_time);
  const nextDay = new Date(date.getTime());
  nextDay.setDate(date.getDate() + 1);
  const prevDay = new Date(date.getTime());
  prevDay.setDate(date.getDate() - 1);
  const cardDetails = {
    cardNumber: data.transaction.account_number?.slice(-4),
    cardType: data.transaction.card_network_type,
    expiryDate: data.transaction.expiry
  }

  const transactionQuery = {
    gatewayName: "WorldPay",
    amount,
    transactionType: "Payment",
    'cardDetails.cardNumber': { $regex: `${cardDetails.cardNumber}$`, $options: 'i' },
    'cardDetails.cardType': { $regex: `${cardDetails.cardType}$`, $options: 'i' },
    createdDate: {
      $gte: prevDay,
      $lte: nextDay
    },
  }

  if (restaurantRefId) {
    transactionQuery['restaurantRefId'] = restaurantRefId
  }

  let transactionIdentifier;
  if (data.transaction.card_network_type === "VISA" && data.transaction.transaction_id) {
    transactionIdentifier = data.transaction.transaction_id?.trim();
    transactionQuery['$or'] = [
      { gatewayResponse: { $regex: `<TransactionIdentifier>\s*.*${transactionIdentifier}.*\s*<\/TransactionIdentifier>`, $options: "i" } },
      { transactionIdentifier: { $regex: transactionIdentifier } }
    ]
  } else if (data.transaction.card_network_type === "MASTERCARD" && data.transaction.banknet_no) {
    transactionIdentifier = data.transaction.banknet_no?.trim();
    transactionQuery['$or'] = [
      { gatewayResponse: { $regex: `<TransactionIdentifier>\s*.*${transactionIdentifier}.*\s*<\/TransactionIdentifier>`, $options: "i" } },
      { transactionIdentifier: { $regex: transactionIdentifier } }
    ]
  } else if (data.transaction.authorization_code) {
    transactionIdentifier = data.transaction.authorization_code?.trim();
    transactionQuery['$or'] = [
      { gatewayResponse: { $regex: `<TransactionIdentifier>\s*.*${transactionIdentifier}.*\s*<\/TransactionIdentifier>`, $options: "i" } },
      { transactionIdentifier: { $regex: transactionIdentifier } }
    ]
  }

  const transactionData = await connection.collection('order_transactions').findOne(transactionQuery);
  if (transactionData) {

    const createMatchedTransactionPromise = connection.collection('matched_order_transaction').updateOne({ refId: transactionData.refId }, {
      $set: {
        refId: transactionData.refId,
        orderRefId: transactionData.orderRefId,
        orderTransactionRefId: transactionData.refId,
        restaurantRefId,
        merchantId: data.merchantId,
        gatewayName: 'WorldPay',
        amount,
        transactionIdentifier: transactionIdentifier?.trim(),
        authorizationCode: data.transaction.authorization_code?.trim(),
        cardDetails,
        settlementData: data.transaction,
        transactionDateTime: date,
        lastModifiedDate: new Date(),
      },
      $setOnInsert: { createdDate: new Date() },
    }, { upsert: true })
    const transactionUpdatePromise = connection.collection('order_transactions').updateOne({ refId: transactionData.refId }, {
      $set: {
        paymentStatus: "Success"
      }
    }).then(result => {
      if (result.matchedCount === 0) {
        console.warn(`order_transactions update matched 0 documents for refId: ${transactionData.refId} — stale read or deleted transaction`);
      }
    })
    const sendMessagePromise = sendMessage(process.env.PRODUCER_EVENT, {
      orderRefId: transactionData.orderRefId,
      transactionRefId: transactionData.refId,
      paymentStatus: "Success",
    })
    await Promise.all([createMatchedTransactionPromise, transactionUpdatePromise, sendMessagePromise])
  } else {
    const createdDate = new Date()
    const unmatchedTransaction = {
      refId: generateUUID(),
      orderRefId: '',
      billId: '',
      amount,
      restaurantRefId,
      merchantId: data.merchantId,
      gatewayName: 'WorldPay',
      transactionIdentifier: transactionIdentifier?.trim(),
      authorizationCode: data.transaction.authorization_code?.trim(),
      refundAmount: 0,
      summary: {},
      cardDetails,
      posSessionRefId: '',
      paymentDeviceId: '',
      createdDate: createdDate,
      lastModifiedDate: createdDate,
      transactionDateTime: date,
      settlementData: data.transaction,
    };
    await connection.collection('unmatched_order_transaction').findOneAndUpdate(
      {
        amount,
        transactionIdentifier: unmatchedTransaction.transactionIdentifier,
        transactionDateTime: date,
        'cardDetails.cardNumber': cardDetails.cardNumber,
        'cardDetails.cardType': cardDetails.cardType,
      },
      { $setOnInsert: unmatchedTransaction },
      { upsert: true }
    );
  }
}

export const handler = async (event, context) => {
  context.callbackWaitsForEmptyEventLoop = false;
  try {
    await Promise.all([connectToDatabase(), connectToKafka()]);
    for (const { body } of event.Records) {
      await mapTransaction(JSON.parse(body))
    }
    return {
      statusCode: 200,
      body: JSON.stringify({ message: 'Transactions processed successfully' }),
    }
  } catch (err) {
    console.error('Error:', err);
    return {
      statusCode: 500,
      body: JSON.stringify({ message: 'Internal Server Error' }),
    };
  }
};
