import {
  connect
} from 'react-redux'

import {
  changeTabIndex
} from '../actions'
import App from '../components/App.jsx'

const mapStateToProps = state => ({
  commandsCount: state.commands.data.length,
  commandsRowsPerPage: state.commands.rowsPerPage,
  sessionsCount: state.sessions.data.length,
  sessionsRowsPerPage: state.sessions.rowsPerPage,
  tabIndex: state.app.tabIndex
})

const mapDispatchToProps = dispatch => ({
  onChangeTabIndex: (event, tabIndex) => dispatch(changeTabIndex(tabIndex))
})

export default connect(mapStateToProps, mapDispatchToProps)(App)
