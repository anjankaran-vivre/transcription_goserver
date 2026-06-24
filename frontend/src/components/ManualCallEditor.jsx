import React, { useState } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { getCallByID, updateCallManually, postCallToZoho } from '../services/api';

export default function ManualCallEditor({ darkMode = true }) {
  const [callID, setCallID] = useState('');
  const [callData, setCallData] = useState(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');
  
  const [editingTranscript, setEditingTranscript] = useState(false);
  const [editingSummary, setEditingSummary] = useState(false);
  
  const [transcript, setTranscript] = useState('');
  const [summary, setSummary] = useState('');
  
  const [updating, setUpdating] = useState(false);
  const [postingToZoho, setPostingToZoho] = useState(false);

  const handleFetchCall = async () => {
    if (!callID.trim()) {
      setError('Please enter a call ID');
      return;
    }

    setLoading(true);
    setError('');
    setSuccess('');

    try {
      const response = await getCallByID(callID);
      if (response.call) {
        setCallData(response.call);
        setTranscript(response.call.transcription || '');
        setSummary(response.call.summary || '');
        setError('');
        setEditingTranscript(false);
        setEditingSummary(false);
      } else {
        setError('Call not found');
        setCallData(null);
      }
    } catch (err) {
      setError(err.response?.data?.error || 'Failed to fetch call');
      setCallData(null);
    } finally {
      setLoading(false);
    }
  };

  const handleUpdateCall = async () => {
    setUpdating(true);
    setError('');
    setSuccess('');

    try {
      const response = await updateCallManually(callID, transcript, summary, false);
      setSuccess('Call updated successfully in database');
      setEditingTranscript(false);
      setEditingSummary(false);
      if (response.call) {
        setCallData(response.call);
      }
    } catch (err) {
      setError(err.response?.data?.error || 'Failed to update call');
    } finally {
      setUpdating(false);
    }
  };

  const handlePostToZoho = async () => {
    setPostingToZoho(true);
    setError('');
    setSuccess('');

    try {
      const response = await postCallToZoho(callID, transcript, summary);
      if (response.success) {
        setSuccess('Successfully posted to Zoho CRM!');
      } else {
        setError(response.error || 'Failed to post to Zoho');
      }
    } catch (err) {
      setError(err.response?.data?.error || 'Failed to post to Zoho');
    } finally {
      setPostingToZoho(false);
    }
  };

  const handleUpdateAndPostToZoho = async () => {
    setUpdating(true);
    setPostingToZoho(true);
    setError('');
    setSuccess('');

    try {
      const response = await updateCallManually(callID, transcript, summary, true);
      if (response.zoho_updated) {
        setSuccess('Call updated in database and posted to Zoho CRM!');
        setEditingTranscript(false);
        setEditingSummary(false);
      } else if (response.db_updated) {
        setSuccess('Call updated in database but failed to post to Zoho: ' + response.zoho_error);
      } else {
        setError('Failed to update call');
      }
    } catch (err) {
      setError(err.response?.data?.error || 'Failed to update and post');
    } finally {
      setUpdating(false);
      setPostingToZoho(false);
    }
  };

  return (
    <motion.div
      className={`rounded-xl shadow-lg p-6 ${
        darkMode ? 'bg-gray-800' : 'bg-white'
      }`}
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
    >
      {/* Header */}
      <div className="flex items-center justify-between mb-6">
        <div className="flex items-center space-x-3">
          <span className="text-2xl">🔧</span>
          <h3
            className={`text-lg font-semibold ${
              darkMode ? 'text-white' : 'text-gray-900'
            }`}
          >
            Manual Call Editor
          </h3>
        </div>
      </div>

      {/* Call ID Input Section */}
      <div className="space-y-4 mb-6">
        <div className="flex space-x-2">
          <input
            type="text"
            placeholder="Enter Call ID"
            value={callID}
            onChange={(e) => setCallID(e.target.value)}
            onKeyPress={(e) => e.key === 'Enter' && handleFetchCall()}
            disabled={loading}
            className={`flex-1 px-4 py-2 rounded-lg border transition ${
              darkMode
                ? 'bg-gray-700 border-gray-600 text-white placeholder-gray-400 focus:border-blue-500'
                : 'bg-white border-gray-300 text-gray-900 placeholder-gray-500 focus:border-blue-500'
            } focus:outline-none`}
          />
          <button
            onClick={handleFetchCall}
            disabled={loading}
            className={`px-4 py-2 rounded-lg font-medium transition ${
              loading
                ? 'bg-gray-600 text-gray-400 cursor-not-allowed'
                : darkMode
                ? 'bg-blue-600 hover:bg-blue-700 text-white'
                : 'bg-blue-500 hover:bg-blue-600 text-white'
            }`}
          >
            {loading ? 'Fetching...' : 'Fetch'}
          </button>
        </div>

        {/* Error and Success Messages */}
        <AnimatePresence>
          {error && (
            <motion.div
              initial={{ opacity: 0, y: -10 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{ opacity: 0, y: -10 }}
              className={`p-3 rounded-lg text-sm ${
                darkMode
                  ? 'bg-red-900 text-red-200'
                  : 'bg-red-100 text-red-800'
              }`}
            >
              ❌ {error}
            </motion.div>
          )}
          {success && (
            <motion.div
              initial={{ opacity: 0, y: -10 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{ opacity: 0, y: -10 }}
              className={`p-3 rounded-lg text-sm ${
                darkMode
                  ? 'bg-green-900 text-green-200'
                  : 'bg-green-100 text-green-800'
              }`}
            >
              ✓ {success}
            </motion.div>
          )}
        </AnimatePresence>
      </div>

      {/* Call Details Display */}
      <AnimatePresence>
        {callData && (
          <motion.div
            initial={{ opacity: 0, y: 10 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: -10 }}
            className="space-y-6"
          >
            {/* Call Info */}
            <div
              className={`p-4 rounded-lg ${
                darkMode ? 'bg-gray-700' : 'bg-gray-100'
              }`}
            >
              <div className="grid grid-cols-2 md:grid-cols-4 gap-4 text-sm">
                <div>
                  <span className={darkMode ? 'text-gray-400' : 'text-gray-600'}>
                    Status:
                  </span>
                  <p
                    className={`font-semibold ${
                      darkMode ? 'text-white' : 'text-gray-900'
                    }`}
                  >
                    {callData.status}
                  </p>
                </div>
                <div>
                  <span className={darkMode ? 'text-gray-400' : 'text-gray-600'}>
                    Duration:
                  </span>
                  <p
                    className={`font-semibold ${
                      darkMode ? 'text-white' : 'text-gray-900'
                    }`}
                  >
                    {callData.duration_sec.toFixed(2)}s
                  </p>
                </div>
                <div>
                  <span className={darkMode ? 'text-gray-400' : 'text-gray-600'}>
                    Quality:
                  </span>
                  <p
                    className={`font-semibold ${
                      darkMode ? 'text-white' : 'text-gray-900'
                    }`}
                  >
                    {callData.audio_quality}
                  </p>
                </div>
                <div>
                  <span className={darkMode ? 'text-gray-400' : 'text-gray-600'}>
                    Words:
                  </span>
                  <p
                    className={`font-semibold ${
                      darkMode ? 'text-white' : 'text-gray-900'
                    }`}
                  >
                    {callData.word_count}
                  </p>
                </div>
              </div>
            </div>

            {/* Transcript Section */}
            <div className="space-y-2">
              <div className="flex items-center justify-between">
                <label
                  className={`font-semibold ${
                    darkMode ? 'text-white' : 'text-gray-900'
                  }`}
                >
                  Transcript
                </label>
                <button
                  onClick={() => setEditingTranscript(!editingTranscript)}
                  className={`text-xs px-3 py-1 rounded transition ${
                    editingTranscript
                      ? darkMode
                        ? 'bg-blue-600 text-white'
                        : 'bg-blue-500 text-white'
                      : darkMode
                      ? 'bg-gray-700 text-blue-400 hover:bg-gray-600'
                      : 'bg-gray-200 text-blue-600 hover:bg-gray-300'
                  }`}
                >
                  {editingTranscript ? '✓ Done' : '✏️ Edit'}
                </button>
              </div>
              <textarea
                value={transcript}
                onChange={(e) => setTranscript(e.target.value)}
                readOnly={!editingTranscript}
                rows={5}
                className={`w-full px-4 py-3 rounded-lg border transition resize-none ${
                  editingTranscript
                    ? darkMode
                      ? 'bg-gray-700 border-blue-500'
                      : 'bg-white border-blue-500'
                    : darkMode
                    ? 'bg-gray-700 border-gray-600'
                    : 'bg-gray-100 border-gray-300'
                } ${darkMode ? 'text-white' : 'text-gray-900'} focus:outline-none`}
              />
            </div>

            {/* Summary Section */}
            <div className="space-y-2">
              <div className="flex items-center justify-between">
                <label
                  className={`font-semibold ${
                    darkMode ? 'text-white' : 'text-gray-900'
                  }`}
                >
                  Summary
                </label>
                <button
                  onClick={() => setEditingSummary(!editingSummary)}
                  className={`text-xs px-3 py-1 rounded transition ${
                    editingSummary
                      ? darkMode
                        ? 'bg-blue-600 text-white'
                        : 'bg-blue-500 text-white'
                      : darkMode
                      ? 'bg-gray-700 text-blue-400 hover:bg-gray-600'
                      : 'bg-gray-200 text-blue-600 hover:bg-gray-300'
                  }`}
                >
                  {editingSummary ? '✓ Done' : '✏️ Edit'}
                </button>
              </div>
              <textarea
                value={summary}
                onChange={(e) => setSummary(e.target.value)}
                readOnly={!editingSummary}
                rows={3}
                className={`w-full px-4 py-3 rounded-lg border transition resize-none ${
                  editingSummary
                    ? darkMode
                      ? 'bg-gray-700 border-blue-500'
                      : 'bg-white border-blue-500'
                    : darkMode
                    ? 'bg-gray-700 border-gray-600'
                    : 'bg-gray-100 border-gray-300'
                } ${darkMode ? 'text-white' : 'text-gray-900'} focus:outline-none`}
              />
            </div>

            {/* Action Buttons */}
            <div className="flex space-x-3 pt-4 border-t border-gray-600">
              <button
                onClick={handleUpdateCall}
                disabled={updating || postingToZoho}
                className={`flex-1 px-4 py-2 rounded-lg font-medium transition ${
                  updating || postingToZoho
                    ? 'bg-gray-600 text-gray-400 cursor-not-allowed'
                    : darkMode
                    ? 'bg-green-600 hover:bg-green-700 text-white'
                    : 'bg-green-500 hover:bg-green-600 text-white'
                }`}
              >
                {updating ? 'Updating...' : '💾 Save to Database'}
              </button>
              <button
                onClick={handlePostToZoho}
                disabled={updating || postingToZoho}
                className={`flex-1 px-4 py-2 rounded-lg font-medium transition ${
                  updating || postingToZoho
                    ? 'bg-gray-600 text-gray-400 cursor-not-allowed'
                    : darkMode
                    ? 'bg-purple-600 hover:bg-purple-700 text-white'
                    : 'bg-purple-500 hover:bg-purple-600 text-white'
                }`}
              >
                {postingToZoho ? 'Posting...' : '📤 Post to Zoho'}
              </button>
              <button
                onClick={handleUpdateAndPostToZoho}
                disabled={updating || postingToZoho}
                className={`flex-1 px-4 py-2 rounded-lg font-medium transition ${
                  updating || postingToZoho
                    ? 'bg-gray-600 text-gray-400 cursor-not-allowed'
                    : darkMode
                    ? 'bg-blue-600 hover:bg-blue-700 text-white'
                    : 'bg-blue-500 hover:bg-blue-600 text-white'
                }`}
              >
                {updating || postingToZoho ? 'Processing...' : '🚀 Save & Post'}
              </button>
            </div>
          </motion.div>
        )}
      </AnimatePresence>

      {!callData && !loading && (
        <div
          className={`text-center py-8 ${
            darkMode ? 'text-gray-400' : 'text-gray-500'
          }`}
        >
          <p className="text-sm">Enter a Call ID and click Fetch to get started</p>
        </div>
      )}
    </motion.div>
  );
}
