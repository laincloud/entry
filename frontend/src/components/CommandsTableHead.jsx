import React from 'react';
import PropTypes from 'prop-types'
import {
  TableCell,
  TableHead,
  TableRow,
  TableSortLabel
} from 'material-ui/Table'
import Tooltip from 'material-ui/Tooltip'

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
    id: 'app',
    numeric: false,
    disablePadding: true,
    label: 'AppName.ProcName.InstanceNo'
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
]

const CommandsTableHead = ({
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

CommandsTableHead.propTypes = {
  orderBy: PropTypes.string.isRequired,
  orderDirection: PropTypes.string.isRequired,
  onSort: PropTypes.func.isRequired
}

export default CommandsTableHead
