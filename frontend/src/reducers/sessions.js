const defaultState = {
  queryStyle: {
    marginTop: '25vh'
  },
  tableStyle: {
    display: 'none'
  },
  data: [],
  orderBy: 'sessionID',
  orderDirection: 'desc',
  page: 0,
  rowsPerPage: 5
}

const sessions = (state = defaultState, action) => {
  switch (action.type) {
    case 'CHANGE_SESSIONS_PAGE':
      return { ...state,
        page: action.page
      }
    case 'CHANGE_SESSIONS_ROWS_PER_PAGE':
      return { ...state,
        rowsPerPage: action.rowsPerPage
      }
    case 'ON_FULFILLED_FETCH_SESSIONS':
      action.response.data.sort(
        (a, b) => b.session_id < a.session_id ? -1 : 1);
      return { ...state,
        data: [...(action.offset === 0) ? [] : state.data,
          ...action.response.data.map(x => ({
            sessionID: x.session_id,
            user: x.user,
            sourceIP: x.source_ip,
            appName: x.app_name,
            procName: x.proc_name,
            instanceNo: x.instance_no,
            nodeIP: x.node_ip,
            status: x.status,
            createdAt: new Date(x.created_at * 1000),
            endedAt: new Date(x.ended_at * 1000)
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
    case 'SORT_SESSIONS':
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
        orderBy: action.orderBy,
        orderDirection: direction
      }
    default:
      return state
  }
};

export default sessions
