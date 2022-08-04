drop table if exists orderInfo, deliveryInfo, paymentInfo, itemInfo;
CREATE TABLE orderInfo
(
  order_uid varchar(50),
  track_number text,
  entry varchar(100),
  locale varchar(20),
  internal_signature text,
  customer_id text,
  delivery_service varchar(100),
  shardkey varchar(20),
  sm_id bigint,
  date_created text,
  oof_shard varchar(20),
  PRIMARY KEY(order_uid)
);
CREATE TABLE deliveryInfo
(
  delivery_id serial,
  name text,
  phone varchar(20),
  zip varchar(20),
  city varchar(100),
  address text,
  region varchar(100),
  email varchar(100),
  order_id varchar(50),
  FOREIGN KEY (order_id) REFERENCES orderInfo (order_uid),
  PRIMARY KEY(delivery_id)
);
CREATE TABLE paymentInfo
(
  payment_id serial,
  transaction text,
  request_id text,
  currency varchar(20),
  provider varchar(100),
  amount integer,
  payment_dt bigint,
  bank varchar(30),
  delivery_cost integer,
  goods_total integer,
  custom_fee integer,
  order_id varchar(50),
  FOREIGN KEY (order_id) REFERENCES orderInfo (order_uid),
  PRIMARY KEY(payment_id)
);
CREATE TABLE itemInfo
(
  item_id serial,
  chrt_id bigint,
  track_number text,
  price integer,
  rid text,
  name text,
  sale integer,
  size varchar(10),
  total_price integer,
  nm_id bigint,
  brand text,
  status integer,
  order_id varchar(50),
  FOREIGN KEY (order_id) REFERENCES orderInfo (order_uid),
  PRIMARY KEY(item_id)
);
