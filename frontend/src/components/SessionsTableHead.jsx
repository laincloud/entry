import React from 'react';
import {
  TableCell,
  TableHead,
  TableRow,
  TableSortLabel
} from 'material-ui/Table'
import Tooltip from 'material-ui/Tooltip'
import PropTypes from 'prop-types'

const columnData = [{
    id: 'sessionID',
    numeric: true,
    disablePadding: false,
    label: 'Session ID'
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
]

const SessionsTableHead = ({
  orderBy,
  orderDirection,
  onSort
}) => (
  <TableHead>
    <TableRow>
      {columnData.map(column => {
        return (
          <TableCell
            key={column.id}
            numeric={column.numeric}
            padding={column.disablePadding ? 'none' : 'default'}
            sortDirection={orderBy === column.id ? orderDirection : false}
          >
            <Tooltip
              title="Sort"
              placement={column.numeric ? 'bottom-end' : 'bottom-start'}
              enterDelay={300}
            >
              <TableSortLabel
                active={orderBy === column.id}
                direction={orderDirection}
                onClick={() => onSort(column.id)}
              >
                {column.label}
              </TableSortLabel>
            </Tooltip>
          </TableCell>
        )
      }, this)}
    </TableRow>
  </TableHead>
)

SessionsTableHead.propTypes = {
  orderBy: PropTypes.string.isRequired,
  orderDirection: PropTypes.string.isRequired,
  onSort: PropTypes.func.isRequired
}

export default SessionsTableHead
