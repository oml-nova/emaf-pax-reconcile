import { S3 } from '@aws-sdk/client-s3';
class S3Service {
  async getFile(params) {
    try {
      const s3 = new S3({ region: 'ap-south-1' });
      let file = await s3.getObject(params);
      return file;
    } catch (err) {
      console.error('Error deleting the file', { error: err, fileName: params })
    }
  }
}

export const StorageService = new S3Service();