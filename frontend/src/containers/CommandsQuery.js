import {
  connect
} from 'react-redux'

import {
  changeFetchCommandsParameter,
  fetchCommands,
  onFulfilledFetchCommands
} from '../actions'
import CommandsQuery from '../components/CommandsQuery.jsx'

const mapStateToProps = state => ({
  appName: state.app.fetchCommandsParameter_appName,
  content: state.app.fetchCommandsParameter_content,
  sessionID: state.app.fetchCommandsParameter_sessionID,
  since: state.app.fetchCommandsParameter_since,
  user: state.app.fetchCommandsParameter_user
})

const mapDispatchToProps = dispatch => ({
  onClick: () =>
    dispatch(fetchCommands(0, offset => response => dispatch(
      onFulfilledFetchCommands(offset)(response)))),
  onChangeTextField: (name, value) => dispatch(changeFetchCommandsParameter(
    name, value))
})

export default connect(mapStateToProps, mapDispatchToProps)(CommandsQuery)
