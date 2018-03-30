import {
  connect
} from 'react-redux'

import {
  changeCommandsPage,
  changeCommandsRowsPerPage,
  fetchCommands,
  onFulfilledFetchCommands
} from '../actions'
import Commands from '../components/Commands.jsx'
import {
  LIMIT
} from '../reducers/myAxios'

const mapStateToProps = state => ({
  queryStyle: state.commands.queryStyle,
  tableStyle: state.commands.tableStyle,
  data: state.commands.data,
  page: state.commands.page,
  rowsPerPage: state.commands.rowsPerPage
})

const mapDispatchToProps = (dispatch, ownProps) => ({
  onChangePage: (event, page) => {
    dispatch(changeCommandsPage(page))
    if ((ownProps.count % LIMIT === 0) && ((page + 1) * ownProps.rowsPerPage >=
        ownProps.count)) {
      dispatch(fetchCommands(ownProps.count, offset => response =>
        dispatch(onFulfilledFetchCommands(offset)(response))))
    }
  },
  onChangeRowsPerPage: event => dispatch(changeCommandsRowsPerPage(event.target
    .value))
})

export default connect(mapStateToProps, mapDispatchToProps)(Commands)
