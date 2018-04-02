import {
  LIMIT,
  get
} from './myAxios'

const params = new URLSearchParams(window.location.search.substring(1))
const fetchSessionsParameter_sessionID = params.get(
  'fetch_sessions_parameter_session_id')
const fetchSessionsParameter_since = params.get(
  'fetch_sessions_parameter_since')
const defaultState = {
  fetchCommandsParameter_appName: '',
  fetchCommandsParameter_content: '',
  fetchCommandsParameter_sessionID: '',
  fetchCommandsParameter_since: new Date(),
  fetchCommandsParameter_user: '',
  fetchSessionsParameter_appName: '',
  fetchSessionsParameter_sessionID: fetchSessionsParameter_sessionID ?
    fetchSessionsParameter_sessionID : '',
  fetchSessionsParameter_since: fetchSessionsParameter_since ?
    new Date(fetchSessionsParameter_since * 1000) : new Date(),
  fetchSessionsParameter_user: '',
  replaySessionID: '',
  tabIndex: 0
}

const app = (state = defaultState, action) => {
  let params = {}
  switch (action.type) {
    case 'CHANGE_FETCH_COMMANDS_PARAMETER':
      return { ...state,
        ['fetchCommandsParameter_' + action.name]: action.value
      }
    case 'CHANGE_FETCH_SESSIONS_PARAMETER':
      return { ...state,
        ['fetchSessionsParameter_' + action.name]: action.value
      }
    case 'CHANGE_REPLAY_SESSION_ID':
      return { ...state,
        replaySessionID: action.sessionID
      }
    case 'CHANGE_TAB_INDEX':
      return { ...state,
        tabIndex: action.tabIndex
      }
    case 'FETCH_COMMANDS':
      params = {
        limit: LIMIT,
        offset: action.offset,
        since: Math.floor(state.fetchCommandsParameter_since / 1000)
      }
      if (state.fetchCommandsParameter_user) {
        params['user'] = '%' + state.fetchCommandsParameter_user + '%';
      }
      if (state.fetchCommandsParameter_appName) {
        params['app_name'] = '%' + state.fetchCommandsParameter_appName + '%';
      }
      if (state.fetchCommandsParameter_sessionID) {
        params['session_id'] = state.fetchCommandsParameter_sessionID;
      }
      if (state.fetchCommandsParameter_content) {
        params['content'] = state.fetchCommandsParameter_content;
      }
      get('/api/commands', action.onFulfilled(action.offset), {
        params: params
      })
      return state
    case 'FETCH_SESSIONS':
      params = {
        limit: LIMIT,
        offset: action.offset,
        since: Math.floor(state.fetchSessionsParameter_since / 1000)
      }
      if (state.fetchSessionsParameter_user) {
        params['user'] = '%' + state.fetchSessionsParameter_user + '%';
      }
      if (state.fetchSessionsParameter_appName) {
        params['app_name'] = '%' + state.fetchSessionsParameter_appName + '%';
      }
      get('/api/sessions', action.onFulfilled(action.offset), {
        params: params
      })
      return state
    case 'SEARCH_COMMANDS_IN_ONE_SESSION':
      return { ...state,
        tabIndex: 1,
        fetchCommandsParameter_sessionID: action.sessionID,
        fetchCommandsParameter_since: action.since
      }
    default:
      return state
  }
}

export default app
