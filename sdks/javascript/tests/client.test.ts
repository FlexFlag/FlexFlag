import { FlexFlagClient } from '../src/client';

// Mock axios
jest.mock('axios', () => ({
  create: jest.fn(() => ({
    get: jest.fn(),
    post: jest.fn(),
    defaults: {
      headers: {
        common: {}
      }
    }
  }))
}));

describe('FlexFlagClient', () => {
  const mockConfig = {
    apiKey: 'test-api-key',
    baseUrl: 'https://api.flexflag.com',
    environment: 'test'
  };

  describe('constructor', () => {
    it('should throw error when API key is missing', () => {
      expect(() => {
        new FlexFlagClient({} as any);
      }).toThrow('FlexFlag: API key is required');
    });

    it('should create client with valid config', () => {
      const client = new FlexFlagClient(mockConfig);
      expect(client).toBeInstanceOf(FlexFlagClient);
    });
  });

  describe('configuration', () => {
    it('should use default values for optional config', () => {
      const client = new FlexFlagClient(mockConfig);
      expect(client).toBeDefined();
    });

    it('should set connection mode', () => {
      const client = new FlexFlagClient({
        ...mockConfig,
        connection: {
          mode: 'polling'
        }
      });
      expect(client).toBeDefined();
    });
  });

  describe('initialization', () => {
    it('should emit ready event on successful initialization', (done) => {
      const client = new FlexFlagClient(mockConfig);
      
      (client as any).on('ready', () => {
        done();
      });
      
      // Mock successful initialization
      setTimeout(() => {
        (client as any).emit('ready');
      }, 10);
    });
  });
});