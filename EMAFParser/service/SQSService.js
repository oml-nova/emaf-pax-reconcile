import { SQSClient, SendMessageBatchCommand } from '@aws-sdk/client-sqs';
class SQSService {
  sqs;
  constructor() {
    this.sqs = new SQSClient({ region: 'ap-south-1', apiVersion: 'v1'})
  }

  async sendBatch(data) {
    try {
      const params = {
        QueueUrl: process.env.DEST_QUEUE_URI,
        Entries: data
      };
      await this.sqs.send(new SendMessageBatchCommand(params));
    } catch (err) {
      console.error('Error Pushing to SQS', JSON.stringify(err))
    }
  }
}

export const QueueService = new SQSService();