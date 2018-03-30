// app
export const changeFetchCommandsParameter = (name, value) => ({
  type: 'CHANGE_FETCH_COMMANDS_PARAMETER',
  name,
  value
})

export const changeFetchSessionsParameter = (name, value) => ({
  type: 'CHANGE_FETCH_SESSIONS_PARAMETER',
  name,
  value
})

export const changeReplaySessionID = sessionID => ({
  type: 'CHANGE_REPLAY_SESSION_ID',
  sessionID
})

export const changeTabIndex = tabIndex => ({
  type: 'CHANGE_TAB_INDEX',
  tabIndex
})

export const fetchCommands = (offset, onFulfilled) => ({
  type: 'FETCH_COMMANDS',
  offset,
  onFulfilled
})

export const fetchSessions = (offset, onFulfilled) => ({
  type: 'FETCH_SESSIONS',
  offset,
  onFulfilled
})

export const searchCommandsInOneSession = (sessionID, since) => ({
  type: 'SEARCH_COMMANDS_IN_ONE_SESSION',
  sessionID,
  since
})

// commands
export const changeCommandsPage = page => ({
  type: 'CHANGE_COMMANDS_PAGE',
  page
})

export const changeCommandsRowsPerPage = rowsPerPage => ({
  type: 'CHANGE_COMMANDS_ROWS_PER_PAGE',
  rowsPerPage
})

export const onFulfilledFetchCommands = offset => response => ({
  type: 'ON_FULFILLED_FETCH_COMMANDS',
  offset,
  response
})

export const sortCommands = orderBy => ({
  type: 'SORT_COMMANDS',
  orderBy
})

// session
export const changeSessionsPage = page => ({
  type: 'CHANGE_SESSIONS_PAGE',
  page
})

export const changeSessionsRowsPerPage = rowsPerPage => ({
  type: 'CHANGE_SESSIONS_ROWS_PER_PAGE',
  rowsPerPage
})

export const onFulfilledFetchSessions = offset => response => ({
  type: 'ON_FULFILLED_FETCH_SESSIONS',
  offset,
  response
})

export const sortSessions = orderBy => ({
  type: 'SORT_SESSIONS',
  orderBy
})
