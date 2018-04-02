const defaultState = {
  queryStyle: {
    marginTop: '25vh'
  },
  tableStyle: {
    display: 'none'
  },
  data: [],
  orderBy: 'commandID',
  orderDirection: 'asc',
  page: 0,
  rowsPerPage: 5
}

const commands = (state = defaultState, action) => {
  switch (action.type) {
    case 'CHANGE_COMMANDS_PAGE':
      return { ...state,
        page: action.page
      }
    case 'CHANGE_COMMANDS_ROWS_PER_PAGE':
      return { ...state,
        rowsPerPage: action.rowsPerPage
      }
    case 'ON_FULFILLED_FETCH_COMMANDS':
      action.response.data.sort(
        (a, b) => a.command_id < b.command_id ? -1 : 1);
      return { ...state,
        data: [...(action.offset === 0) ? [] : state.data,
          ...action.response.data.map(x => ({
            commandID: x.command_id,
            user: x.user,
            appName: x.app_name,
            procName: x.proc_name,
            instanceNo: x.instance_no,
            content: x.content,
            sessionID: x.session_id,
            createdAt: new Date(x.created_at * 1000)
          }))
        ],
        page: (action.offset === 0) ? 0 : state.page,
        queryStyle: {
          marginTop: '0vh'
        },
        tableStyle: {
          display: 'block'
        }
      }
    case 'SORT_COMMANDS':
      let direction = 'desc'
      let orderBy = action.orderBy
      if (state.orderBy === orderBy && state.orderDirection === 'desc') {
        direction = 'asc'
      }

      let data = [...state.data]
      if (direction === 'desc') {
        data.sort((a, b) => (b[orderBy] < a[orderBy] ? -1 : 1))
      } else {
        data.sort((a, b) => (a[orderBy] < b[orderBy] ? -1 : 1))
      }
      return { ...state,
        data: data,
        orderBy: orderBy,
        orderDirection: direction
      }
    default:
      return state
  }
};

export default commands
