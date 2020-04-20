
/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET NAMES utf8 */;
/*!50503 SET NAMES utf8mb4 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;


CREATE DATABASE IF NOT EXISTS `devplayg` /*!40100 DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci */;
USE `devplayg`;

CREATE TABLE IF NOT EXISTS `log` (
                                     `log_id` bigint(20) NOT NULL AUTO_INCREMENT,
                                     `date` datetime NOT NULL,
                                     `device_id` bigint(20) NOT NULL,
                                     `event_type` int(11) NOT NULL,
                                     `uuid` varchar(256) COLLATE utf8mb4_unicode_ci NOT NULL,
                                     `flag` int(11) NOT NULL DEFAULT 0,
                                     `created` datetime NOT NULL DEFAULT current_timestamp(),
                                     PRIMARY KEY (`log_id`),
                                     KEY `ix_log_date` (`date`),
                                     KEY `ix_log_date_deviceId` (`date`,`device_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

/*!40101 SET SQL_MODE=IFNULL(@OLD_SQL_MODE, '') */;
/*!40014 SET FOREIGN_KEY_CHECKS=IF(@OLD_FOREIGN_KEY_CHECKS IS NULL, 1, @OLD_FOREIGN_KEY_CHECKS) */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
