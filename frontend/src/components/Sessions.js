import React from 'react';
import classNames from 'classnames';
import PropTypes from 'prop-types';
import {
  withStyles
} from 'material-ui/styles';
import Table, {
  TableBody,
  TableCell,
  TableFooter,
  TableHead,
  TablePagination,
  TableRow,
  TableSortLabel
} from 'material-ui/Table';
import Toolbar from 'material-ui/Toolbar';
import Typography from 'material-ui/Typography';
import Paper from 'material-ui/Paper';
import Tooltip from 'material-ui/Tooltip';
import {
  lighten
} from 'material-ui/styles/colorManipulator';
import MenuItem from 'material-ui/Menu/MenuItem';
import TextField from 'material-ui/TextField';
import Button from 'material-ui/Button';
import IconButton from 'material-ui/IconButton';
import ReplayIcon from 'material-ui-icons/Replay';
import SearchIcon from 'material-ui-icons/Search';
import {
  format
} from 'date-fns';

import MyDateTimePicker from './MyDateTimePicker.jsx';
import {
  get
} from '../MyAxios.jsx';

const LIMIT = 200;
const SESSION_TYPES = ['enter', 'attach'];

const queryStyles = theme => ({
  container: {
    display: 'flex',
    flexWrap: 'wrap',
    marginBottom: theme.spacing.unit * 15,
    justifyContent: 'center'
  },
  textField: {
    marginLeft: theme.spacing.unit,
    marginRight: theme.spacing.unit,
    minWidth: 200
  },
  menu: {
    width: 200
  },
  button: {
    margin: theme.spacing.unit * 3,
  }
})

class Query extends React.Component {
  render() {
    const {
      classes,
      sessionType,
      onSessionTypeChange,
      user,
      onUserChange,
      appName,
      onAppNameChange,
      since,
      onSinceChange,
      onClick
    } = this.props;

    return (
      <form className={classes.container}>
        <TextField
          id="sessionType"
          select
          label="Session Type"
          className={classes.textField}
          value={sessionType}
          onChange={onSessionTypeChange}
          SelectProps={{
            MenuProps: {
              className: classes.menu
            }
          }}
          margin="normal"
        >
          {SESSION_TYPES.map(option => (
            <MenuItem key={option} value={option}>
              {option}
            </MenuItem>
          ))}
        </TextField>

        <TextField
          id="user"
          label="User"
          className={classes.textField}
          value={user}
          onChange={onUserChange}
          margin="normal"
        />

        <TextField
          id="appName"
          label="App Name"
          className={classes.textField}
          value={appName}
          onChange={onAppNameChange}
          margin="normal"
        />

        <MyDateTimePicker
          value={since}
          onChange={onSinceChange}
        />

        <Button
          variant="raised"
          color="primary"
          className={classes.button}
          onClick={onClick}
        >
          Query
        </Button>
      </form>
    );
  }
}

Query.propTypes = {
  classes: PropTypes.object.isRequired,
  sessionType: PropTypes.string.isRequired,
  onSessionTypeChange: PropTypes.func.isRequired,
  user: PropTypes.string.isRequired,
  onUserChange: PropTypes.func.isRequired,
  appName: PropTypes.string.isRequired,
  onAppNameChange: PropTypes.func.isRequired,
  since: PropTypes.object.isRequired,
  onSinceChange: PropTypes.func.isRequired,
  onClick: PropTypes.func.isRequired
};

Query = withStyles(queryStyles)(Query);

const columnData = [{
    id: 'sessionID',
    numeric: true,
    disablePadding: false,
    label: 'Session ID'
  },
  {
    id: 'sessionType',
    numeric: false,
    disablePadding: true,
    label: 'Session Type'
  },
  {
    id: 'user',
    numeric: false,
    disablePadding: true,
    label: 'User'
  },
  {
    id: 'sourceIP',
    numeric: false,
    disablePadding: true,
    label: 'Source IP'
  },
  {
    id: 'app',
    numeric: false,
    disablePadding: true,
    label: 'AppName.ProcName.InstanceNo'
  },
  {
    id: 'nodeIP',
    numeric: false,
    disablePadding: true,
    label: 'Node IP'
  },
  {
    id: 'status',
    numeric: false,
    disablePadding: true,
    label: 'Status'
  },
  {
    id: 'createdAt',
    numeric: false,
    disablePadding: true,
    label: 'Created At'
  },
  {
    id: 'endedAt',
    numeric: false,
    disablePadding: true,
    label: 'Ended At'
  },
  {
    id: 'inspect',
    numeric: false,
    disablePadding: true,
    label: 'Inspect'
  }
];

class EnhancedTableHead extends React.Component {
  createSortHandler = property => event => {
    this.props.onRequestSort(event, property);
  };

  render() {
    const {
      order,
      orderBy,
    } = this.props;

    return (
      <TableHead>
        <TableRow>
          {columnData.map(column => {
            return (
              <TableCell
                key={column.id}
                numeric={column.numeric}
                padding={column.disablePadding ? 'none' : 'default'}
                sortDirection={orderBy === column.id ? order : false}
              >
                <Tooltip
                  title="Sort"
                  placement={column.numeric ? 'bottom-end' : 'bottom-start'}
                  enterDelay={300}
                >
                  <TableSortLabel
                    active={orderBy === column.id}
                    direction={order}
                    onClick={this.createSortHandler(column.id)}
                  >
                    {column.label}
                  </TableSortLabel>
                </Tooltip>
              </TableCell>
            );
          }, this)}
        </TableRow>
      </TableHead>
    );
  }
}

EnhancedTableHead.propTypes = {
  onRequestSort: PropTypes.func.isRequired,
  order: PropTypes.string.isRequired,
  orderBy: PropTypes.string.isRequired,
};

const toolbarStyles = theme => ({
  root: {
    paddingRight: theme.spacing.unit
  },
  highlight: theme.palette.type === 'light' ? {
    color: theme.palette.secondary.dark,
    backgroundColor: lighten(theme.palette.secondary.light, 0.4)
  } : {
    color: lighten(theme.palette.secondary.light, 0.4),
    backgroundColor: theme.palette.secondary.dark
  },
  spacer: {
    flex: '0 0 auto'
  },
  actions: {
    color: theme.palette.text.secondary
  },
  title: {
    flex: '0 0 auto'
  }
})

let EnhancedTableToolbar = props => {
  const {
    classes
  } = props;

  return (
    <Toolbar
      className={classNames(classes.root)}
    >
      <div className={classes.title}>
        <Typography variant="title">Sessions</Typography>
      </div>
    </Toolbar>
  )
};

EnhancedTableToolbar.propTypes = {
  classes: PropTypes.object.isRequired
};

EnhancedTableToolbar = withStyles(toolbarStyles)(EnhancedTableToolbar);

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

class Sessions extends React.Component {
  constructor(props, context) {
    super(props, context);

    this.state = {
      sessionType: 'enter',
      user: '',
      appName: '',
      since: new Date(),
      order: 'desc',
      orderBy: 'sessionID',
      data: [],
      page: 0,
      rowsPerPage: 5,
      queryStyle: {
        marginTop: '30vh'
      },
      tableStyle: {
        display: 'none'
      }
    };
  }

  handleTextFieldChange = name => event => {
    this.setState({
      [name]: event.target.value
    });
  }

  handleSinceChange = since => {
    this.setState({
      since: since
    })
  }

  handleRequestSort = (event, property) => {
    const orderBy = property;
    let order = 'desc';

    if (this.state.orderBy === property && this.state.order === 'desc') {
      order = 'asc';
    }

    const data = order === 'desc' ?
      this.state.data.sort((a, b) => (b[orderBy] < a[orderBy] ? -1 : 1)) :
      this.state.data.sort((a, b) => (a[orderBy] < b[orderBy] ? -1 : 1));

    this.setState({
      data,
      order,
      orderBy
    });
  }

  loadData = response => {
    response.data.sort((a, b) => b.session_id < a.session_id ? -1 :
      1);
    let data = response.data.map(x => ({
      sessionID: x.session_id,
      sessionType: x.session_type,
      user: x.user,
      sourceIP: x.source_ip,
      appName: x.app_name,
      procName: x.proc_name,
      instanceNo: x.instance_no,
      nodeIP: x.node_ip,
      status: x.status,
      createdAt: new Date(x.created_at * 1000),
      endedAt: new Date(x.ended_at * 1000)
    }));

    this.setState({
      data: data,
      queryStyle: {
        marginTop: '0vh'
      },
      tableStyle: {
        display: 'block'
      },
      page: 0
    });
  }

  loadMoreData = response => {
    response.data.sort((a, b) => b.session_id < a.session_id ? -1 :
      1);
    let newData = response.data.map(x => ({
      sessionID: x.session_id,
      sessionType: x.session_type,
      user: x.user,
      sourceIP: x.source_ip,
      appName: x.app_name,
      procName: x.proc_name,
      instanceNo: x.instance_no,
      nodeIP: x.node_ip,
      status: x.status,
      createdAt: new Date(x.created_at * 1000),
      endedAt: new Date(x.ended_at * 1000)
    }));
    this.setState({
      data: this.state.data.concat(newData)
    });
  }

  handleChangePage = (event, page) => {
    this.setState({
      page
    });

    let count = this.state.data.length;
    if ((count % LIMIT === 0) && ((page + 1) * this.state.rowsPerPage >=
        count)) {
      let params = {
        limit: LIMIT,
        offset: count,
        since: Math.floor(this.state.since / 1000)
      }
      if (this.state.user) {
        params['user'] = this.state.user;
      }
      if (this.state.appName) {
        params['app_name'] = this.state.appName;
      }
      get('/api/sessions', this.loadMoreData, {
        params: params
      });
    }
  }

  handleChangeRowsPerPage = event => {
    this.setState({
      rowsPerPage: event.target.value
    });
  }

  handleQuery = () => {
    let params = {
      limit: LIMIT,
      offset: 0,
      since: Math.floor(this.state.since / 1000)
    }

    if (this.state.user) {
      params['user'] = this.state.user;
    }

    if (this.state.appName) {
      params['app_name'] = this.state.appName;
    }

    get('/api/sessions', this.loadData, {
      params: params
    });
  }

  render() {
    const {
      classes,
      onSearchCommands
    } = this.props;
    const {
      sessionType,
      user,
      appName,
      since,
      data,
      order,
      orderBy,
      rowsPerPage,
      page,
      queryStyle,
      tableStyle
    } = this.state;
    const emptyRows = rowsPerPage - Math.min(rowsPerPage, data.length -
      page * rowsPerPage);

    return (
      <div>
        <div style={queryStyle}>
          <Query
            sessionType={sessionType}
            onSessionTypeChange={this.handleTextFieldChange('sessionType')}
            user={user}
            onUserChange={this.handleTextFieldChange('user')}
            appName={appName}
            onAppNameChange={this.handleTextFieldChange('appName')}
            since={since}
            onSinceChange={this.handleSinceChange}
            onClick={this.handleQuery}
            colSpan={12}
          />
        </div>

        <div style={tableStyle}>
          <Paper className={classes.root}>
            <EnhancedTableToolbar />

            <div className={classes.tableWrapper}>
              <Table className={classes.table}>
                <EnhancedTableHead
                  order={order}
                  orderBy={orderBy}
                  onRequestSort={this.handleRequestSort}
                />

                <TableBody>
                  {data.slice(page * rowsPerPage, page * rowsPerPage + rowsPerPage).map(n => {
                    return (
                      <TableRow
                        hover
                        tabIndex={-1}
                        key={n.sessionID}
                      >
                        <TableCell numeric>{n.sessionID}</TableCell>
                        <TableCell padding="none">{n.sessionType}</TableCell>
                        <TableCell padding="none">{n.user}</TableCell>
                        <TableCell padding="none">{n.sourceIP}</TableCell>
                        <TableCell padding="none">{n.appName}.{n.procName}.{n.instanceNo}</TableCell>
                        <TableCell padding="none">{n.nodeIP}</TableCell>
                        <TableCell padding="none">{n.status}</TableCell>
                        <TableCell padding="none">{format(n.createdAt, 'YYYY-MM-DD HH:mm:ss')}</TableCell>
                        <TableCell padding="none">{format(n.endedAt, 'YYYY-MM-DD HH:mm:ss')}</TableCell>
                        <TableCell padding="none">
                          <Tooltip title="Replay">
                            <IconButton className={classes.button} aria-label="Replay">
                              <ReplayIcon />
                            </IconButton>
                          </Tooltip>
                          <Tooltip title="Search Commands">
                            <IconButton
                              className={classes.button}
                              aria-label="Search"
                              onClick={() => onSearchCommands(n.sessionID.toString(), n.createdAt)}
                            >
                              <SearchIcon />
                            </IconButton>
                          </Tooltip>
                        </TableCell>
                      </TableRow>
                    );
                  })}
                  {emptyRows > 0 && (
                    <TableRow style={{ height: 49 * emptyRows }}>
                      <TableCell colSpan={12} />
                    </TableRow>
                  )}
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
                      onChangePage={this.handleChangePage}
                      onChangeRowsPerPage={this.handleChangeRowsPerPage}
                    />
                  </TableRow>
                </TableFooter>
              </Table>
            </div>
          </Paper>
        </div>
      </div>
    );
  }
}

Sessions.propTypes = {
  classes: PropTypes.object.isRequired,
  onSearchCommands: PropTypes.func.isRequired
};

export default withStyles(styles)(Sessions);
