import {
  connect
} from 'react-redux'

import {
  changeFetchSessionsParameter,
  fetchSessions,
  onFulfilledFetchSessions
} from '../actions'
import SessionsQuery from '../components/SessionsQuery.jsx'

const mapStateToProps = state => ({
  appName: state.app.fetchSessionsParameter_appName,
  sessionID: state.app.fetchSessionsParameter_sessionID,
  since: state.app.fetchSessionsParameter_since,
  user: state.app.fetchSessionsParameter_user
})

const mapDispatchToProps = dispatch => ({
  onClick: () =>
    dispatch(fetchSessions(0, offset => response => dispatch(
      onFulfilledFetchSessions(offset)(response)))),
  onChangeTextField: (name, value) => dispatch(changeFetchSessionsParameter(
    name, value))
})

export default connect(mapStateToProps, mapDispatchToProps)(SessionsQuery)
