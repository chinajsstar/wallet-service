# exadmin api-admin
# Base images
FROM centos
# Base Pkg
RUN yum install -y epel-release.noarch && \
    yum install -y vim wget net-tools iproute telnet git mysql-devel redis  tree sudo psmisc && \
    yum install -y nginx supervisor && \
    yum clean all

# Install Ngninx
# RUN yum install -y nginx
# RUN mkdir -p /usr/share/nginx/html/documents
# ADD configuration/exadmin_api/nginx_conf.d/nginx-default.conf /etc/nginx/conf.d/nginx-default.conf
# ADD configuration/exadmin_api/nginx_conf.d/nginx.conf /etc/nginx/nginx.conf

# Install Supervisord
# RUN yum install -y supervisor
# ADD configuration/bastionpay/supervisord.d/supervisord-cobank_srv.ini  /etc/supervisord.d/supervisord-cobank_srv.ini
# ADD configuration/bastionpay/supervisord.d/supervisord.conf /etc/supervisord.d/supervisord.conf
ADD configuration/bastionpay/supervisord.d/ /etc/supervisord.d/

# Install api-admin
# ADD cobank_srv /opt/
ADD configuration/bastionpay/BastionPay/ /root/BastionPay/
ADD configuration/bastionpay/blockchain_server /root/blockchain_server/
ADD bin/linux/ /opt/
## EXPOSE
EXPOSE  8081 8082 8077
CMD ["/usr/bin/supervisord","-c","/etc/supervisord.d/supervisord.conf"]
