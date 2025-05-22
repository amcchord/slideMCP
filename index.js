/**
 * Slide API Model Configuration Profile (MCP)
 * This module provides an interface for LLMs to interact with the Slide API
 */

const axios = require('axios');
require('dotenv').config();

class SlideClient {
  constructor(apiKey, baseURL = 'https://api.slide.tech', version = 'v1') {
    this.apiKey = apiKey;
    this.baseURL = baseURL;
    this.version = version;
    
    this.client = axios.create({
      baseURL: `${this.baseURL}/${this.version}`,
      headers: {
        'Authorization': `Bearer ${this.apiKey}`,
        'Content-Type': 'application/json',
      }
    });
  }

  // Helper method for GET requests
  async get(endpoint, params = {}) {
    try {
      const response = await this.client.get(endpoint, { params });
      return response.data;
    } catch (error) {
      this._handleError(error);
    }
  }

  // Helper method for POST requests
  async post(endpoint, data = {}) {
    try {
      const response = await this.client.post(endpoint, data);
      return response.data;
    } catch (error) {
      this._handleError(error);
    }
  }

  // Helper method for PATCH requests
  async patch(endpoint, data = {}) {
    try {
      const response = await this.client.patch(endpoint, data);
      return response.data;
    } catch (error) {
      this._handleError(error);
    }
  }

  // Helper method for DELETE requests
  async delete(endpoint) {
    try {
      const response = await this.client.delete(endpoint);
      return response.data;
    } catch (error) {
      this._handleError(error);
    }
  }

  // Error handler
  _handleError(error) {
    if (error.response) {
      // The request was made and the server responded with a status code
      // that falls out of the range of 2xx
      throw new Error(`API Error: ${JSON.stringify(error.response.data)}`);
    } else if (error.request) {
      // The request was made but no response was received
      throw new Error('No response received from API');
    } else {
      // Something happened in setting up the request that triggered an Error
      throw new Error(`Error: ${error.message}`);
    }
  }

  // Devices API
  async getDevices(params = {}) {
    return this.get('/device', params);
  }

  async getDevice(deviceId) {
    return this.get(`/device/${deviceId}`);
  }

  async updateDevice(deviceId, data) {
    return this.patch(`/device/${deviceId}`, data);
  }

  // Agents API
  async getAgents(params = {}) {
    return this.get('/agent', params);
  }

  async getAgent(agentId) {
    return this.get(`/agent/${agentId}`);
  }

  async updateAgent(agentId, data) {
    return this.patch(`/agent/${agentId}`, data);
  }

  async createAgentPair(data) {
    return this.post('/agent', data);
  }

  async pairAgent(data) {
    return this.post('/agent/pair', data);
  }

  // Backups API
  async getBackups(params = {}) {
    return this.get('/backup', params);
  }

  async getBackup(backupId) {
    return this.get(`/backup/${backupId}`);
  }

  async startBackup(data) {
    return this.post('/backup', data);
  }

  // Snapshots API
  async getSnapshots(params = {}) {
    return this.get('/snapshot', params);
  }

  async getSnapshot(snapshotId) {
    return this.get(`/snapshot/${snapshotId}`);
  }

  // File Restores API
  async getFileRestores(params = {}) {
    return this.get('/restore/file', params);
  }

  async getFileRestore(fileRestoreId) {
    return this.get(`/restore/file/${fileRestoreId}`);
  }

  async createFileRestore(data) {
    return this.post('/restore/file', data);
  }

  async deleteFileRestore(fileRestoreId) {
    return this.delete(`/restore/file/${fileRestoreId}`);
  }

  async browseFileRestore(fileRestoreId, params = {}) {
    return this.get(`/restore/file/${fileRestoreId}/browse`, params);
  }

  // Image Exports API
  async getImageExports(params = {}) {
    return this.get('/restore/image', params);
  }

  async getImageExport(imageExportId) {
    return this.get(`/restore/image/${imageExportId}`);
  }

  async createImageExport(data) {
    return this.post('/restore/image', data);
  }

  async deleteImageExport(imageExportId) {
    return this.delete(`/restore/image/${imageExportId}`);
  }

  async browseImageExport(imageExportId, params = {}) {
    return this.get(`/restore/image/${imageExportId}/browse`, params);
  }

  // Virtual Machines API
  async getVirtualMachines(params = {}) {
    return this.get('/restore/virt', params);
  }

  async getVirtualMachine(virtId) {
    return this.get(`/restore/virt/${virtId}`);
  }

  async createVirtualMachine(data) {
    return this.post('/restore/virt', data);
  }

  async updateVirtualMachine(virtId, data) {
    return this.patch(`/restore/virt/${virtId}`, data);
  }

  async deleteVirtualMachine(virtId) {
    return this.delete(`/restore/virt/${virtId}`);
  }

  // Networks API
  async getNetworks(params = {}) {
    return this.get('/network', params);
  }

  async getNetwork(networkId) {
    return this.get(`/network/${networkId}`);
  }

  async createNetwork(data) {
    return this.post('/network', data);
  }

  async updateNetwork(networkId, data) {
    return this.patch(`/network/${networkId}`, data);
  }

  async deleteNetwork(networkId) {
    return this.delete(`/network/${networkId}`);
  }

  // Network Port Forwards API
  async createNetworkPortForward(networkId, data) {
    return this.post(`/network/${networkId}/port-forwards`, data);
  }

  async deleteNetworkPortForward(networkId, data) {
    return this.delete(`/network/${networkId}/port-forwards`, data);
  }

  // Network WireGuard Peers API
  async createNetworkWGPeer(networkId, data) {
    return this.post(`/network/${networkId}/wg-peers`, data);
  }

  async updateNetworkWGPeer(networkId, data) {
    return this.patch(`/network/${networkId}/wg-peers`, data);
  }

  async deleteNetworkWGPeer(networkId, data) {
    return this.delete(`/network/${networkId}/wg-peers`, data);
  }

  // Users API
  async getUsers(params = {}) {
    return this.get('/user', params);
  }

  async getUser(userId) {
    return this.get(`/user/${userId}`);
  }

  // Alerts API
  async getAlerts(params = {}) {
    return this.get('/alert', params);
  }

  async getAlert(alertId) {
    return this.get(`/alert/${alertId}`);
  }

  async updateAlert(alertId, data) {
    return this.patch(`/alert/${alertId}`, data);
  }

  // Account API
  async getAccounts(params = {}) {
    return this.get('/account', params);
  }

  async getAccount(accountId) {
    return this.get(`/account/${accountId}`);
  }

  async updateAccount(accountId, data) {
    return this.patch(`/account/${accountId}`, data);
  }
}

// Factory function to create a new SlideClient instance
function createClient(apiKey) {
  if (!apiKey) {
    throw new Error('API key is required');
  }
  return new SlideClient(apiKey);
}

module.exports = {
  SlideClient,
  createClient
}; 