import React from 'react';
import PropTypes from 'prop-types';
import {
  withStyles
} from 'material-ui/styles';
import Table, {
  TableBody,
  TableCell,
  TableFooter,
  TablePagination,
  TableRow,
} from 'material-ui/Table';
import Paper from 'material-ui/Paper';
import Tooltip from 'material-ui/Tooltip';
import IconButton from 'material-ui/IconButton';
import ReplayIcon from 'material-ui-icons/Replay';
import SearchIcon from 'material-ui-icons/Search';
import {
  format
} from 'date-fns';

import SessionsQuery from '../containers/SessionsQuery'
import SessionsTableHead from '../containers/SessionsTableHead'
import SessionsTableToolbar from '../components/SessionsTableToolbar.jsx'

const styles = theme => ({
  button: {
    margin: theme.spacing.unit,
  },
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
});

const Sessions = ({
  classes,
  queryStyle,
  tableStyle,
  data,
  page,
  rowsPerPage,
  onChangePage,
  onChangeRowsPerPage,
  onReplay,
  onSearchCommands
}) => (
  <div>
    <div style={queryStyle}>
      <SessionsQuery colSpan={12} />
    </div>

    <div style={tableStyle}>
      <Paper className={classes.root}>
        <SessionsTableToolbar />

        <div className={classes.tableWrapper}>
          <Table className={classes.table}>
            <SessionsTableHead />

            <TableBody>
              {data.slice(page * rowsPerPage, page * rowsPerPage + rowsPerPage).map(n => {
                return (
                  <TableRow
                    hover
                    tabIndex={-1}
                    key={n.sessionID}
                  >
                    <TableCell numeric>{n.sessionID}</TableCell>
                    <TableCell padding="none">{n.user}</TableCell>
                    <TableCell padding="none">{n.sourceIP}</TableCell>
                    <TableCell padding="none">{n.appName}.{n.procName}.{n.instanceNo}</TableCell>
                    <TableCell padding="none">{n.nodeIP}</TableCell>
                    <TableCell padding="none">{n.status}</TableCell>
                    <TableCell padding="none">{format(n.createdAt, 'YYYY-MM-DD HH:mm:ss')}</TableCell>
                    <TableCell padding="none">{format(n.endedAt, 'YYYY-MM-DD HH:mm:ss')}</TableCell>
                    <TableCell padding="none">
                      <Tooltip title="Search Commands">
                        <IconButton
                          className={classes.button}
                          aria-label="Search"
                          onClick={() => onSearchCommands(n.sessionID.toString(), n.createdAt)}
                        >
                          <SearchIcon />
                        </IconButton>
                      </Tooltip>

                      <Tooltip title="Replay">
                        <IconButton
                          className={classes.button}
                          aria-label="Replay"
                          onClick={() => onReplay(n.sessionID)}
                        >
                          <ReplayIcon />
                        </IconButton>
                      </Tooltip>
                    </TableCell>
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

Sessions.propTypes = {
  classes: PropTypes.object.isRequired,
  queryStyle: PropTypes.object.isRequired,
  tableStyle: PropTypes.object.isRequired,
  data: PropTypes.array.isRequired,
  page: PropTypes.number.isRequired,
  rowsPerPage: PropTypes.number.isRequired,
  onChangePage: PropTypes.func.isRequired,
  onChangeRowsPerPage: PropTypes.func.isRequired,
  onReplay: PropTypes.func.isRequired,
  onSearchCommands: PropTypes.func.isRequired
};

export default withStyles(styles)(Sessions);
