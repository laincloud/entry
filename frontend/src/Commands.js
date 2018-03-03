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
import TextField from 'material-ui/TextField';
import Button from 'material-ui/Button';
import axios from 'axios';
import {
  format
} from 'date-fns';

import MyDateTimePicker from './MyDateTimePicker.jsx';

const LIMIT = 100;

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

var myAxios = axios.create({
  headers: {
    'Accept': 'application/vnd.laincloud.entry.v1+json'
  }
});

class Query extends React.Component {
  render() {
    const {
      classes,
      user,
      onUserChange,
      appName,
      onAppNameChange,
      sessionID,
      onSessionIDChange,
      content,
      onContentChange,
      since,
      onSinceChange,
      onClick
    } = this.props;

    return (
      <form className={classes.container}>
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

        <TextField
          id="content"
          label="Content(MySQL LIKE)"
          className={classes.textField}
          value={content}
          onChange={onContentChange}
          margin="normal"
        />

        <TextField
          id="sessionID"
          label="Session ID"
          className={classes.textField}
          value={sessionID}
          onChange={onSessionIDChange}
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
  user: PropTypes.string.isRequired,
  onUserChange: PropTypes.func.isRequired,
  appName: PropTypes.string.isRequired,
  onAppNameChange: PropTypes.func.isRequired,
  content: PropTypes.object.isRequired,
  onContentChange: PropTypes.func.isRequired,
  sessionID: PropTypes.string.isRequired,
  onSessionIDChange: PropTypes.func.isRequired,
  since: PropTypes.object.isRequired,
  onSinceChange: PropTypes.func.isRequired,
  onClick: PropTypes.func.isRequired
};

Query = withStyles(queryStyles)(Query);

const columnData = [{
    id: 'commandID',
    numeric: true,
    disablePadding: false,
    label: 'Command ID'
  },
  {
    id: 'user',
    numeric: false,
    disablePadding: true,
    label: 'User'
  },
  {
    id: 'appName',
    numeric: false,
    disablePadding: true,
    label: 'App Name'
  },
  {
    id: 'procName',
    numeric: false,
    disablePadding: true,
    label: 'Proc Name'
  },
  {
    id: 'instanceNo',
    numeric: true,
    disablePadding: false,
    label: 'Instance No'
  },
  {
    id: 'content',
    numeric: false,
    disablePadding: true,
    label: 'Content'
  },
  {
    id: 'sessionID',
    numeric: true,
    disablePadding: false,
    label: 'Session ID'
  },
  {
    id: 'createdAt',
    numeric: false,
    disablePadding: true,
    label: 'Created At'
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
        <Typography variant="title">Commands</Typography>
      </div>
    </Toolbar>
  )
};

EnhancedTableToolbar.propTypes = {
  classes: PropTypes.object.isRequired
};

EnhancedTableToolbar = withStyles(toolbarStyles)(EnhancedTableToolbar);

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
});

class Commands extends React.Component {
  constructor(props, context) {
    super(props, context);

    this.state = {
      user: '',
      appName: '',
      since: new Date(),
      order: 'desc',
      orderBy: 'commandID',
      data: [],
      page: 0,
      rowsPerPage: 5,
      queryStyle: {
        marginTop: '30vh'
      },
      tableStyle: {
        display: 'none'
      },
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
      if (this.state.sessionID) {
        params['session_id'] = this.state.sessionID;
      }
      if (this.state.content) {
        params['content'] = this.state.content;
      }
      myAxios.get('/api/commands', {
          params: params
        })
        .then(response => {
          console.info(response);
          let data = this.state.data.slice();
          response.data.sort((a, b) => b.command_id < a.command_id ? -1 :
            1);
          response.data.forEach(item => {
            data.push({
              commandID: item.command_id,
              user: item.user,
              appName: item.app_name,
              procName: item.proc_name,
              instanceNo: item.instance_no,
              content: item.content,
              sessionID: item.session_id,
              createdAt: format(new Date(item.created_at * 1000),
                'YYYY-MM-DD HH:mm:ss')
            });
          });
          this.setState({
            data: data
          });
        })
        .catch(err => {
          console.error(err);
        });
    }
  }

  handleChangeRowsPerPage = event => {
    this.setState({
      rowsPerPage: event.target.value
    });
  }

  handleQuery = () => {
    this.setState({
      queryStyle: {
        marginTop: '0vh'
      },
      tableStyle: {
        display: 'block'
      },
      page: 0
    });

    console.info(this.state);
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
    if (this.state.sessionID) {
      params['session_id'] = this.state.sessionID;
    }
    if (this.state.content) {
      params['content'] = this.state.content;
    }
    myAxios.get('/api/commands', {
        params: params
      })
      .then(response => {
        console.info(response);
        let data = [];
        response.data.sort((a, b) => b.command_id < a.command_id ? -1 : 1);
        response.data.forEach(item => {
          data.push({
            commandID: item.command_id,
            user: item.user,
            appName: item.app_name,
            procName: item.proc_name,
            instanceNo: item.instance_no,
            content: item.content,
            sessionID: item.session_id,
            createdAt: format(new Date(item.created_at * 1000),
              'YYYY-MM-DD HH:mm:ss')
          });
        });
        this.setState({
          data: data
        });
      })
      .catch(err => {
        console.error(err);
      });
  }

  render() {
    const {
      classes
    } = this.props;
    const {
      user,
      appName,
      sessionID,
      content,
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
            user={user}
            onUserChange={this.handleTextFieldChange('user')}
            appName={appName}
            onAppNameChange={this.handleTextFieldChange('appName')}
            content={content}
            onContentChange={this.handleTextFieldChange('content')}
            sessionID={sessionID}
            onSessionIDChange={this.handleTextFieldChange('sessionID')}
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
                        key={n.commandID}
                      >
                        <TableCell numeric>{n.commandID}</TableCell>
                        <TableCell padding="none">{n.user}</TableCell>
                        <TableCell padding="none">{n.appName}</TableCell>
                        <TableCell padding="none">{n.procName}</TableCell>
                        <TableCell numeric>{n.instanceNo}</TableCell>
                        <TableCell padding="none">{n.content}</TableCell>
                        <TableCell numeric>{n.sessionID}</TableCell>
                        <TableCell padding="none">{n.createdAt}</TableCell>
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

Commands.propTypes = {
  classes: PropTypes.object.isRequired
};

export default withStyles(styles)(Commands);
