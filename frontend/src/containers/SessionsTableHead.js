import {
  connect
} from 'react-redux'

import {
  sortSessions
} from '../actions'
import SessionsTableHead from '../components/SessionsTableHead.jsx'

const mapStateToProps = state => ({
  orderBy: state.sessions.orderBy,
  orderDirection: state.sessions.orderDirection
})

const mapDispatchToProps = dispatch => ({
  onSort: orderBy => dispatch(sortSessions(orderBy))
})

export default connect(mapStateToProps, mapDispatchToProps)(SessionsTableHead)
