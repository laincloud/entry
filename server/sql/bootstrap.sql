CREATE TABLE `sessions` (
`session_id` bigint(20) NOT NULL AUTO_INCREMENT,
`user` varchar(255) DEFAULT NULL,
`source_ip` varchar(255) DEFAULT NULL,
`app_name` varchar(255) DEFAULT NULL,
`proc_name` varchar(255) DEFAULT NULL,
`instance_no` varchar(255) DEFAULT NULL,
`container_id` varchar(255) DEFAULT NULL,
`node_ip` varchar(255) DEFAULT NULL,
`status` varchar(255) DEFAULT NULL,
`ended_at` timestamp NULL DEFAULT NULL,
`created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
`updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
PRIMARY KEY (`session_id`),
KEY `idx_sessions_user` (`user`(191)),
KEY `idx_sessions_source_ip` (`source_ip`(191)),
KEY `idx_sessions_app_name` (`app_name`(191))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `commands` (
`command_id` bigint(20) NOT NULL AUTO_INCREMENT,
`session_id` bigint(20) DEFAULT NULL,
`user` varchar(255) DEFAULT NULL,
`content` varchar(1024) DEFAULT NULL,
`created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
PRIMARY KEY (`command_id`),
KEY `idx_commands_user` (`user`(191)),
FOREIGN KEY (`session_id`) REFERENCES `sessions`(`session_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
