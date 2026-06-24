import axios from 'axios';

// Determine the API base URL
const getBaseURL = () => {
  // In development, use localhost
  if (window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1') {
    return 'http://localhost:5050';
  }
  
  // In production, use the same domain but assume backend is on same host
  // This works if there's a proxy or backend is on same domain
  const protocol = window.location.protocol;
  const host = window.location.host;
  return `${protocol}//${host}`;
};

const api = axios.create({
  baseURL: getBaseURL(),
  timeout: 10000,
});

// Log the API base URL for debugging
console.log('API Base URL:', getBaseURL());

export const getStatus = async () => {
  const response = await api.get('/api/status');
  return response.data;
};

export const getMetrics = async () => {
  const response = await api.get('/api/metrics');
  return response.data;
};

export const getCalls = async (limit = 50) => {
  const response = await api.get(`/api/calls?limit=${limit}`);
  return response.data;
};

export const getLogs = async (limit = 100) => {
  const response = await api.get(`/api/logs?limit=${limit}`);
  return response.data;
};

export const getProcessing = async () => {
  const response = await api.get('/api/processing');
  return response.data;
};

export const restartServer = async () => {
  const response = await api.post('/admin/restart');
  return response.data;
};

export const stopServer = async () => {
  const response = await api.post('/admin/stop');
  return response.data;
};

export const clearQueue = async () => {
  const response = await api.post('/admin/clear-queue');
  return response.data;
};

export const testEmail = async () => {
  const response = await api.post('/admin/test-email');
  return response.data;
};

// Manual call editing endpoints
export const getCallByID = async (callID) => {
  const response = await api.get(`/api/call/${callID}`);
  return response.data;
};

export const updateCallManually = async (callID, transcription, summary, postToZoho = false) => {
  const response = await api.post(`/api/call/${callID}/update-manual`, {
    transcription,
    summary,
    post_to_zoho: postToZoho,
  });
  return response.data;
};

export const postCallToZoho = async (callID, transcription, summary) => {
  const response = await api.post(`/api/call/${callID}/post-to-zoho`, {
    transcription,
    summary,
  });
  return response.data;
};

export const fetchCallFromZoho = async (callID) => {
  const response = await api.post(`/api/call/${callID}/fetch-from-zoho`);
  return response.data;
};

export default api;