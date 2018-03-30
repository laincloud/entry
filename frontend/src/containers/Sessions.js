import {
  connect
} from 'react-redux'

import {
  changeReplaySessionID,
  changeSessionsPage,
  changeSessionsRowsPerPage,
  changeTabIndex,
  fetchSessions,
  onFulfilledFetchSessions,
  searchCommandsInOneSession
} from '../actions'
import Sessions from '../components/Sessions.jsx'
import {
  LIMIT
} from '../reducers/myAxios'

const mapStateToProps = state => ({
  queryStyle: state.sessions.queryStyle,
  tableStyle: state.sessions.tableStyle,
  data: state.sessions.data,
  page: state.sessions.page,
  rowsPerPage: state.sessions.rowsPerPage
})

const mapDispatchToProps = (dispatch, ownProps) => ({
  onChangePage: (event, page) => {
    dispatch(changeSessionsPage(page))
    if ((ownProps.count % LIMIT === 0) && ((page + 1) * ownProps.rowsPerPage >=
        ownProps.count)) {
      dispatch(fetchSessions(ownProps.count, offset => response =>
        dispatch(onFulfilledFetchSessions(offset)(response))))
    }
  },
  onChangeRowsPerPage: event => dispatch(changeSessionsRowsPerPage(event.target
    .value)),
  onReplay: sessionID => {
    dispatch(changeReplaySessionID(sessionID))
    dispatch(changeTabIndex(2))
  },
  onSearchCommands: (sessionID, since) => dispatch(
    searchCommandsInOneSession(sessionID, since))
})

export default connect(mapStateToProps, mapDispatchToProps)(Sessions)
