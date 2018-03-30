import React from 'react'
import PropTypes from 'prop-types'
import {
  withStyles
} from 'material-ui/styles'
import Table, {
  TableBody,
  TableCell,
  TableFooter,
  TablePagination,
  TableRow,
} from 'material-ui/Table'
import Paper from 'material-ui/Paper'
import {
  format
} from 'date-fns'

import CommandsQuery from '../containers/CommandsQuery'
import CommandsTableHead from '../containers/CommandsTableHead'
import CommandsTableToolbar from './CommandsTableToolbar.jsx'

const styles = theme => ({
  root: {
    width: '100%',
    marginTop: theme.spacing.unit * 3
  },
  table: {
    minWidth: 800
  },
  tableWrapper: {
    overflowX: 'auto'
  }
})

const Commands = ({
  classes,
  queryStyle,
  tableStyle,
  data,
  page,
  rowsPerPage,
  onChangePage,
  onChangeRowsPerPage
}) => (
  <div>
    <div style={queryStyle}>
      <CommandsQuery colSpan={12} />
    </div>

    <div style={tableStyle}>
      <Paper className={classes.root}>
        <CommandsTableToolbar />

        <div className={classes.tableWrapper}>
          <Table className={classes.table}>
            <CommandsTableHead />

            <TableBody>
              {data.slice(page * rowsPerPage, page * rowsPerPage + rowsPerPage).map(n => {
                return (
                  <TableRow
                    hover
                    tabIndex={-1}
                    key={n.commandID}
                  >
                    <TableCell numeric>{n.commandID}</TableCell>
                    <TableCell padding="none">{n.user}</TableCell>
                    <TableCell padding="none">{n.appName}.{n.procName}.{n.instanceNo}</TableCell>
                    <TableCell padding="none">{n.content}</TableCell>
                    <TableCell numeric>{n.sessionID}</TableCell>
                    <TableCell padding="none">{format(n.createdAt, 'YYYY-MM-DD HH:mm:ss')}</TableCell>
                  </TableRow>
                );
              })}
            </TableBody>

            <TableFooter>
              <TableRow>
                <TablePagination
                  colSpan={12}
                  count={data.length}
                  rowsPerPage={rowsPerPage}
                  rowsPerPageOptions={[5, 10, 25, 50, 100]}
                  page={page}
                  backIconButtonProps={{
                    'aria-label': 'Previous Page'
                  }}
                  nextIconButtonProps={{
                    'aria-label': 'Next Page'
                  }}
                  onChangePage={onChangePage}
                  onChangeRowsPerPage={onChangeRowsPerPage}
                />
              </TableRow>
            </TableFooter>
          </Table>
        </div>
      </Paper>
    </div>
  </div>
)

Commands.propTypes = {
  classes: PropTypes.object.isRequired,
  queryStyle: PropTypes.object.isRequired,
  tableStyle: PropTypes.object.isRequired,
  data: PropTypes.array.isRequired,
  page: PropTypes.number.isRequired,
  rowsPerPage: PropTypes.number.isRequired,
  onChangePage: PropTypes.func.isRequired,
  onChangeRowsPerPage: PropTypes.func.isRequired
};

export default withStyles(styles)(Commands);
