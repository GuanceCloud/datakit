use cloudcare_shrine;
DELETE FROM `order` WHERE id = "a21756672bd843c084b945f6";
INSERT INTO `order` (id,customer_id,customer_name,commodity_package_id,commodity_package_name,price,discount_rate,total_price,total_payment,team_id,`status`,contract_name,
  contract_amount,paid_amount,invoiced_payment,collected_payment,create_account_id,update_account_id,create_time,update_time )
VALUES
  ("a21756672bd843c084b945f6","a21756672bd843c084b945f7","上海天下大业有限公司","驻云cms","cloudcare+",30000,"100.00%",30000,30000,"team-ggq8bfWshdqtEDXtBEEe1k","Enabled","2019cms合同",
    30000,10000,30000,20000,"acnt-n8CHDoursQNapJXCd9ALjE","5c9cfdc65ae8200006946d75","2019-10-12 10:00:00","2019-10-16 12:00:00" );
SELECT  *  FROM `order` WHERE id = "a21756672bd843c084b945f6";