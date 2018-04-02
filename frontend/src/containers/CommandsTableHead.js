import {
  connect
} from 'react-redux'

import {
  sortCommands
} from '../actions'
import CommandsTableHead from '../components/CommandsTableHead.jsx'

const mapStateToProps = state => ({
  orderBy: state.commands.orderBy,
  orderDirection: state.commands.orderDirection
})

const mapDispatchToProps = dispatch => ({
  onSort: orderBy => dispatch(sortCommands(orderBy))
})

export default connect(mapStateToProps, mapDispatchToProps)(CommandsTableHead)
