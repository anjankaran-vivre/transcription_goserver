import React, { useCallback, useEffect, useState } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import {
  Database,
  Download,
  Edit3,
  ExternalLink,
  Play,
  RefreshCw,
  Save,
  Search,
  Send,
  Wrench,
  X,
} from 'lucide-react';
import {
  clearCallAudioCache,
  downloadCallAudio,
  fetchCallFromZoho,
  getCallByID,
  postCallToZoho,
  resendCallForTranscription,
  updateCallManually,
} from '../services/api';

export default function ManualCallEditor({ darkMode = true }) {
  const [callID, setCallID] = useState('');
  const [callData, setCallData] = useState(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');

  const [editingTranscript, setEditingTranscript] = useState(false);
  const [editingSummary, setEditingSummary] = useState(false);

  const [rawTranscript, setRawTranscript] = useState('');
  const [transcript, setTranscript] = useState('');
  const [summary, setSummary] = useState('');

  const [updating, setUpdating] = useState(false);
  const [postingToZoho, setPostingToZoho] = useState(false);
  const [fetchingFromZoho, setFetchingFromZoho] = useState(false);
  const [audioDownloading, setAudioDownloading] = useState(false);
  const [audioUrl, setAudioUrl] = useState('');
  const [activeCacheCallID, setActiveCacheCallID] = useState('');
  const [resending, setResending] = useState(false);

  const inputCallID = callID.trim();
  const callURL = callData?.call_url || '';

  const clearCachedAudio = useCallback(async (clearState = true) => {
    const cacheCallID = activeCacheCallID || inputCallID;
    if (!cacheCallID) return;

    try {
      await clearCallAudioCache(cacheCallID);
    } catch {
      // Best effort cleanup; temp files are also replaced on the next download.
    } finally {
      if (clearState) {
        setAudioUrl('');
        setActiveCacheCallID('');
      }
    }
  }, [activeCacheCallID, inputCallID]);

  useEffect(() => {
    return () => {
      if (activeCacheCallID) {
        clearCallAudioCache(activeCacheCallID).catch(() => {});
      }
    };
  }, [activeCacheCallID]);

  const resetEditor = () => {
    setCallData(null);
    setRawTranscript('');
    setTranscript('');
    setSummary('');
    setEditingTranscript(false);
    setEditingSummary(false);
  };

  const applyCallData = (data) => {
    setCallData(data);
    setRawTranscript(data.raw_transcription || '');
    setTranscript(data.transcription || '');
    setSummary(data.summary || '');
    setError('');
    setEditingTranscript(false);
    setEditingSummary(false);
  };

  const downloadAudioForCall = async (targetCallID, targetCallURL, successMessage = 'Audio downloaded') => {
    if (!targetCallID || !targetCallURL) {
      setError('Call recording URL is missing');
      return false;
    }

    setAudioDownloading(true);
    try {
      const response = await downloadCallAudio(targetCallID, targetCallURL);
      setAudioUrl(response.audio_url || '');
      setActiveCacheCallID(targetCallID);
      setSuccess(successMessage);
      return true;
    } catch (err) {
      setAudioUrl('');
      setError(err.response?.data?.error || 'Failed to download audio');
      return false;
    } finally {
      setAudioDownloading(false);
    }
  };

  const handleFetchCall = async () => {
    if (!inputCallID) {
      setError('Please enter a call ID');
      return;
    }

    await clearCachedAudio();
    setLoading(true);
    setError('');
    setSuccess('');

    try {
      const response = await getCallByID(inputCallID);
      if (response.call) {
        applyCallData(response.call);
      } else {
        setError('Call not found');
        resetEditor();
      }
    } catch (err) {
      setError(err.response?.data?.error || 'Failed to fetch call');
      resetEditor();
    } finally {
      setLoading(false);
    }
  };

  const handleFetchFromZoho = async () => {
    if (!inputCallID) {
      setError('Please enter a call ID');
      return;
    }

    await clearCachedAudio();
    setFetchingFromZoho(true);
    setError('');
    setSuccess('');

    try {
      const response = await fetchCallFromZoho(inputCallID);
      if (response.call_id) {
        applyCallData(response);
        if (response.call_url) {
          const downloaded = await downloadAudioForCall(
            inputCallID,
            response.call_url,
            'Call fetched from Zoho CRM and audio is ready to play',
          );
          if (!downloaded) {
            setSuccess('Call fetched from Zoho CRM, but audio download failed');
          }
        } else {
          setSuccess('Call fetched from Zoho CRM, but no recording URL was found');
        }
      } else {
        setError(response.error || 'Failed to fetch from Zoho');
        resetEditor();
      }
    } catch (err) {
      setError(err.response?.data?.error || 'Failed to fetch from Zoho');
      resetEditor();
    } finally {
      setFetchingFromZoho(false);
    }
  };

  const handleDownloadAudio = async () => {
    setError('');
    setSuccess('');
    await downloadAudioForCall(inputCallID, callURL);
  };

  const handleResendTranscription = async () => {
    if (!inputCallID || !callURL) {
      setError('Call recording URL is missing');
      return;
    }

    setResending(true);
    setError('');
    setSuccess('');

    try {
      const response = await resendCallForTranscription(inputCallID, callURL);
      setRawTranscript(response.raw_transcription || '');
      setTranscript(response.transcription || '');
      setSummary(response.summary || '');
      setCallData((current) => ({
        ...current,
        raw_transcription: response.raw_transcription || '',
        transcription: response.transcription || '',
        summary: response.summary || '',
        status: 'manual_transcription',
      }));
      setSuccess(response.message || 'AI transcription refreshed');
    } catch (err) {
      setError(err.response?.data?.error || 'Failed to resend for transcription');
    } finally {
      setResending(false);
    }
  };

  const handleUpdateCall = async () => {
    setUpdating(true);
    setError('');
    setSuccess('');

    try {
      await updateCallManually(inputCallID, transcript, summary, false);
      setSuccess('Call updated successfully in database');
      setEditingTranscript(false);
      setEditingSummary(false);
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
      const response = await postCallToZoho(inputCallID, transcript, summary);
      if (response.success) {
        setSuccess('Successfully posted to Zoho CRM');
        await clearCachedAudio();
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
      const response = await updateCallManually(inputCallID, transcript, summary, true);
      if (response.zoho_updated) {
        setSuccess('Call updated in database and posted to Zoho CRM');
        setEditingTranscript(false);
        setEditingSummary(false);
        await clearCachedAudio();
      } else if (response.db_updated) {
        setSuccess(`Call updated in database but failed to post to Zoho: ${response.zoho_error}`);
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

  const handleCancel = async () => {
    await clearCachedAudio();
    resetEditor();
    setError('');
    setSuccess('');
  };

  const buttonBase = 'inline-flex items-center justify-center gap-2 px-4 py-2 rounded-lg font-medium transition';
  const disabled = loading || fetchingFromZoho || updating || postingToZoho || audioDownloading || resending;

  return (
    <motion.div
      className={`rounded-xl shadow-lg p-6 ${darkMode ? 'bg-gray-800' : 'bg-white'}`}
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
    >
      <div className="flex items-center justify-between mb-6">
        <div className="flex items-center gap-3">
          <Wrench className={darkMode ? 'text-blue-300' : 'text-blue-600'} size={24} />
          <h3 className={`text-lg font-semibold ${darkMode ? 'text-white' : 'text-gray-900'}`}>
            Manual Call Editor
          </h3>
        </div>
      </div>

      <div className="space-y-4 mb-6">
        <div className="flex flex-col lg:flex-row gap-2">
          <input
            type="text"
            placeholder="Enter Call ID"
            value={callID}
            onChange={(e) => setCallID(e.target.value)}
            onKeyDown={(e) => e.key === 'Enter' && handleFetchCall()}
            disabled={disabled}
            className={`flex-1 px-4 py-2 rounded-lg border transition ${
              darkMode
                ? 'bg-gray-700 border-gray-600 text-white placeholder-gray-400 focus:border-blue-500'
                : 'bg-white border-gray-300 text-gray-900 placeholder-gray-500 focus:border-blue-500'
            } focus:outline-none`}
          />
          <button
            onClick={handleFetchCall}
            disabled={disabled}
            className={`${buttonBase} ${
              disabled
                ? 'bg-gray-600 text-gray-400 cursor-not-allowed'
                : darkMode
                  ? 'bg-blue-600 hover:bg-blue-700 text-white'
                  : 'bg-blue-500 hover:bg-blue-600 text-white'
            }`}
            title="Fetch from database"
          >
            <Database size={16} />
            {loading ? 'Fetching...' : 'From DB'}
          </button>
          <button
            onClick={handleFetchFromZoho}
            disabled={disabled}
            className={`${buttonBase} whitespace-nowrap ${
              disabled
                ? 'bg-gray-600 text-gray-400 cursor-not-allowed'
                : darkMode
                  ? 'bg-purple-600 hover:bg-purple-700 text-white'
                  : 'bg-purple-500 hover:bg-purple-600 text-white'
            }`}
            title="Fetch from Zoho"
          >
            <Search size={16} />
            {fetchingFromZoho ? 'Fetching...' : 'From Zoho'}
          </button>
        </div>

        <AnimatePresence>
          {error && (
            <motion.div
              initial={{ opacity: 0, y: -10 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{ opacity: 0, y: -10 }}
              className={`p-3 rounded-lg text-sm ${darkMode ? 'bg-red-900 text-red-200' : 'bg-red-100 text-red-800'}`}
            >
              {error}
            </motion.div>
          )}
          {success && (
            <motion.div
              initial={{ opacity: 0, y: -10 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{ opacity: 0, y: -10 }}
              className={`p-3 rounded-lg text-sm ${darkMode ? 'bg-green-900 text-green-200' : 'bg-green-100 text-green-800'}`}
            >
              {success}
            </motion.div>
          )}
        </AnimatePresence>
      </div>

      <AnimatePresence>
        {callData && (
          <motion.div
            initial={{ opacity: 0, y: 10 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: -10 }}
            className="space-y-6"
          >
            <div className={`p-4 rounded-lg ${darkMode ? 'bg-gray-700' : 'bg-gray-100'}`}>
              <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3 mb-3">
                <span className={`w-fit text-xs px-2 py-1 rounded ${
                  callData.status === 'fetched_from_zoho'
                    ? 'bg-purple-600 text-white'
                    : 'bg-blue-600 text-white'
                }`}>
                  {callData.status === 'fetched_from_zoho' ? 'From Zoho CRM' : 'From Database'}
                </span>
                {callURL && (
                  <a
                    href={callURL}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="inline-flex items-center gap-1 text-xs text-blue-400 hover:underline"
                  >
                    Recording Link
                    <ExternalLink size={13} />
                  </a>
                )}
              </div>
              <div className="grid grid-cols-2 md:grid-cols-4 gap-4 text-sm">
                <InfoItem darkMode={darkMode} label="Status" value={callData.status} />
                <InfoItem darkMode={darkMode} label="Duration" value={`${callData.duration_sec?.toFixed(2) || callData.duration || 'N/A'}s`} />
                <InfoItem darkMode={darkMode} label="Quality" value={callData.audio_quality || 'N/A'} />
                <InfoItem darkMode={darkMode} label="Words" value={callData.word_count || 'N/A'} />
              </div>
            </div>

            {callURL && (
              <div className={`p-4 rounded-lg space-y-3 ${darkMode ? 'bg-gray-700' : 'bg-gray-100'}`}>
                <div className="flex flex-col md:flex-row gap-2">
                  <button
                    onClick={handleDownloadAudio}
                    disabled={disabled}
                    className={`${buttonBase} ${
                      disabled
                        ? 'bg-gray-600 text-gray-400 cursor-not-allowed'
                        : darkMode
                          ? 'bg-slate-600 hover:bg-slate-500 text-white'
                          : 'bg-slate-200 hover:bg-slate-300 text-slate-900'
                    }`}
                    title="Download audio"
                  >
                    <Download size={16} />
                    {audioDownloading ? 'Downloading...' : 'Download Audio'}
                  </button>
                  <button
                    onClick={handleResendTranscription}
                    disabled={disabled}
                    className={`${buttonBase} ${
                      disabled
                        ? 'bg-gray-600 text-gray-400 cursor-not-allowed'
                        : darkMode
                          ? 'bg-amber-600 hover:bg-amber-700 text-white'
                          : 'bg-amber-500 hover:bg-amber-600 text-white'
                    }`}
                    title="Resend for transcription"
                  >
                    <RefreshCw size={16} />
                    {resending ? 'Queuing...' : 'Retrigger AI'}
                  </button>
                </div>
                {audioUrl && (
                  <div className="flex items-center gap-3">
                    <Play size={18} className={darkMode ? 'text-gray-300' : 'text-gray-600'} />
                    <audio controls src={audioUrl} className="w-full h-10" />
                  </div>
                )}
              </div>
            )}

            <TranscriptBox
              darkMode={darkMode}
              label="Raw Transcript"
              value={rawTranscript}
              onChange={setRawTranscript}
              readOnly
              rows={4}
            />

            <TranscriptBox
              darkMode={darkMode}
              label="English Transcript"
              value={transcript}
              onChange={setTranscript}
              readOnly={!editingTranscript}
              rows={5}
              action={
                <EditButton
                  darkMode={darkMode}
                  active={editingTranscript}
                  onClick={() => setEditingTranscript(!editingTranscript)}
                />
              }
            />

            <TranscriptBox
              darkMode={darkMode}
              label="Summary"
              value={summary}
              onChange={setSummary}
              readOnly={!editingSummary}
              rows={3}
              action={
                <EditButton
                  darkMode={darkMode}
                  active={editingSummary}
                  onClick={() => setEditingSummary(!editingSummary)}
                />
              }
            />

            <div className="grid grid-cols-1 md:grid-cols-4 gap-3 pt-4 border-t border-gray-600">
              <button
                onClick={handleUpdateCall}
                disabled={disabled}
                className={`${buttonBase} ${
                  disabled
                    ? 'bg-gray-600 text-gray-400 cursor-not-allowed'
                    : darkMode
                      ? 'bg-green-600 hover:bg-green-700 text-white'
                      : 'bg-green-500 hover:bg-green-600 text-white'
                }`}
              >
                <Save size={16} />
                {updating ? 'Updating...' : 'Save DB'}
              </button>
              <button
                onClick={handlePostToZoho}
                disabled={disabled}
                className={`${buttonBase} ${
                  disabled
                    ? 'bg-gray-600 text-gray-400 cursor-not-allowed'
                    : darkMode
                      ? 'bg-purple-600 hover:bg-purple-700 text-white'
                      : 'bg-purple-500 hover:bg-purple-600 text-white'
                }`}
              >
                <Send size={16} />
                {postingToZoho ? 'Posting...' : 'Post Zoho'}
              </button>
              <button
                onClick={handleUpdateAndPostToZoho}
                disabled={disabled}
                className={`${buttonBase} ${
                  disabled
                    ? 'bg-gray-600 text-gray-400 cursor-not-allowed'
                    : darkMode
                      ? 'bg-blue-600 hover:bg-blue-700 text-white'
                      : 'bg-blue-500 hover:bg-blue-600 text-white'
                }`}
              >
                <Send size={16} />
                {updating || postingToZoho ? 'Processing...' : 'Save & Post'}
              </button>
              <button
                onClick={handleCancel}
                disabled={disabled}
                className={`${buttonBase} ${
                  disabled
                    ? 'bg-gray-600 text-gray-400 cursor-not-allowed'
                    : darkMode
                      ? 'bg-gray-600 hover:bg-gray-500 text-white'
                      : 'bg-gray-200 hover:bg-gray-300 text-gray-900'
                }`}
              >
                <X size={16} />
                Cancel
              </button>
            </div>
          </motion.div>
        )}
      </AnimatePresence>

      {!callData && !loading && (
        <div className={`text-center py-8 ${darkMode ? 'text-gray-400' : 'text-gray-500'}`}>
          <p className="text-sm">Enter a Call ID and click Fetch to get started</p>
        </div>
      )}
    </motion.div>
  );
}

function InfoItem({ darkMode, label, value }) {
  return (
    <div>
      <span className={darkMode ? 'text-gray-400' : 'text-gray-600'}>{label}:</span>
      <p className={`font-semibold break-words ${darkMode ? 'text-white' : 'text-gray-900'}`}>
        {value}
      </p>
    </div>
  );
}

function EditButton({ darkMode, active, onClick }) {
  return (
    <button
      onClick={onClick}
      className={`inline-flex items-center gap-1 text-xs px-3 py-1 rounded transition ${
        active
          ? darkMode
            ? 'bg-blue-600 text-white'
            : 'bg-blue-500 text-white'
          : darkMode
            ? 'bg-gray-700 text-blue-400 hover:bg-gray-600'
            : 'bg-gray-200 text-blue-600 hover:bg-gray-300'
      }`}
      title={active ? 'Finish editing' : 'Edit'}
    >
      <Edit3 size={13} />
      {active ? 'Done' : 'Edit'}
    </button>
  );
}

function TranscriptBox({ darkMode, label, value, onChange, readOnly, rows, action }) {
  return (
    <div className="space-y-2">
      <div className="flex items-center justify-between">
        <label className={`font-semibold ${darkMode ? 'text-white' : 'text-gray-900'}`}>
          {label}
        </label>
        {action}
      </div>
      <textarea
        value={value}
        onChange={(e) => onChange(e.target.value)}
        readOnly={readOnly}
        rows={rows}
        className={`w-full px-4 py-3 rounded-lg border transition resize-none ${
          !readOnly
            ? darkMode
              ? 'bg-gray-700 border-blue-500'
              : 'bg-white border-blue-500'
            : darkMode
              ? 'bg-gray-700 border-gray-600'
              : 'bg-gray-100 border-gray-300'
        } ${darkMode ? 'text-white' : 'text-gray-900'} focus:outline-none`}
      />
    </div>
  );
}
