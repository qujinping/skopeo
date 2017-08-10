FROM mid/builder-go-17-centos7

USER root
RUN yum update -y && yum install -y glib2-devel ostree-devel btrfs-progs-devel gpgme-devel libassuan2-devel device-mapper-devel && yum clean all 

USER 1001
