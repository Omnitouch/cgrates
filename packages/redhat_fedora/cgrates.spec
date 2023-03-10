%global version 0.10.0
%global git_commit %(echo $gitLastCommit)
%global releaseTag %(echo $rpmTag)

%define debug_package  %{nil}
%global _logdir	       /var/log/%name
%global _spooldir      /var/spool/%name
%global _libdir	       /var/lib/%name

Name:           cgrates
Version:        %{version}
Release:        %{releaseTag}
Summary:        Carrier Grade Real-time Charging System
License:        GPLv3
URL:            https://github.com/Omnitouch/cgrates
Source0:        https://github.com/Omnitouch/cgrates/archive/%{git_commit}.tar.gz

%if 0%{?fedora} > 16 || 0%{?rhel} > 6
Requires(pre): shadow-utils
Requires(post): systemd
Requires(preun): systemd
Requires(postun): systemd
%else
Requires(post): chkconfig
Requires(preun):chkconfig
Requires(preun):initscripts
%endif

%description
CGRateS is a very fast and easy scalable real-time charging system for Telecom environments.

%prep
%setup -q -n %{name}-%{version} -c
mkdir -p src/github.com/cgrates
ln -sf ../../../%{name}-%{git_commit} src/github.com/Omnitouch/cgrates

%pre
getent group %{name} >/dev/null || groupadd -r %{name}
getent passwd %{name} >/dev/null || \
useradd -r -g %{name} -d %{_localstatedir}/run/%{name} -s /sbin/nologin \
-c "CGRateS" %{name} 2>/dev/null || :

%post
%if 0%{?fedora} > 16 || 0%{?rhel} > 6
if [ $1 -eq 1 ] ; then
	# Initial installation
	/bin/systemctl daemon-reload >/dev/null 2>&1 || :
fi
%else
/sbin/chkconfig --add %{name}
%endif
/bin/chown -R %{name}:%{name} %{_logdir}
/bin/chown -R %{name}:%{name} %{_spooldir}
/bin/chown -R %{name}:%{name} %{_libdir}

%preun
%if 0%{?fedora} > 16 || 0%{?rhel} > 6
if [ $1 -eq 0 ] ; then
	# Package removal, not upgrade
	/bin/systemctl --no-reload disable %{name}.service > /dev/null 2>&1 || :
	/bin/systemctl stop %{name}.service > /dev/null 2>&1 || :
fi
%else
if [ $1 = 0 ]; then
	/sbin/service %{name} stop > /dev/null 2>&1
	/sbin/chkconfig --del %{name}
fi
%endif

%build
export GOPATH=$RPM_BUILD_DIR/%{name}-%{version}
cd $RPM_BUILD_DIR/%{name}-%{version}/src/github.com/Omnitouch/cgrates
./build.sh

%install
rm -rf $RPM_BUILD_ROOT
mkdir -p $RPM_BUILD_ROOT%{_datarootdir}/%{name}
cp -rpf src/github.com/Omnitouch/cgrates/data/* $RPM_BUILD_ROOT%{_datarootdir}/%{name}
install -D -m 0644 -p src/github.com/Omnitouch/cgrates/data/conf/%{name}/%{name}.json $RPM_BUILD_ROOT%{_sysconfdir}/%{name}/%{name}.json
install -D -m 0755 -p bin/cgr-console $RPM_BUILD_ROOT%{_bindir}/cgr-console
install -D -m 0755 -p bin/cgr-engine $RPM_BUILD_ROOT%{_bindir}/cgr-engine
install -D -m 0755 -p bin/cgr-loader $RPM_BUILD_ROOT%{_bindir}/cgr-loader
install -D -m 0755 -p bin/cgr-tester $RPM_BUILD_ROOT%{_bindir}/cgr-tester
install -D -m 0755 -p bin/cgr-migrator $RPM_BUILD_ROOT%{_bindir}/cgr-migrator
mkdir -p $RPM_BUILD_ROOT%{_logdir}/cdre/csv
mkdir -p $RPM_BUILD_ROOT%{_logdir}/cdre/fwv
mkdir -p $RPM_BUILD_ROOT%{_spooldir}/cdre/csv
mkdir -p $RPM_BUILD_ROOT%{_spooldir}/cdre/fwv
mkdir -p $RPM_BUILD_ROOT%{_spooldir}/tpe
mkdir -p $RPM_BUILD_ROOT%{_spooldir}/failed_posts
mkdir -p $RPM_BUILD_ROOT%{_libdir}/history
mkdir -p $RPM_BUILD_ROOT%{_libdir}/cache_dump
install -D -m 0644 -p src/github.com/Omnitouch/cgrates/packages/redhat_fedora/%{name}.options $RPM_BUILD_ROOT%{_sysconfdir}/sysconfig/%{name}
%if 0%{?fedora} > 16 || 0%{?rhel} > 6
	install -D -m 0644 -p src/github.com/Omnitouch/cgrates/packages/redhat_fedora/%{name}.service $RPM_BUILD_ROOT%{_unitdir}/%{name}.service
%else
	install -D -m 0755 -p src/github.com/Omnitouch/cgrates/packages/redhat_fedora/%{name}.init $RPM_BUILD_ROOT%{_initrddir}/%{name}
%endif

%files
%defattr(-,root,root,-)
%{_datarootdir}/%{name}/*
%{_bindir}/*
%config(noreplace) %{_sysconfdir}/%{name}/%{name}.json
%{_logdir}/*
%{_spooldir}/*
%{_libdir}/*
%{_sysconfdir}/sysconfig/%{name}
%if 0%{?fedora} > 16 || 0%{?rhel} > 6
	%{_unitdir}/%{name}.service
%else
	%{_initrddir}/%{name}
%endif

%clean
sudo rm -rf $RPM_BUILD_DIR/%{name}-%{version}
sudo rm -rf $RPM_BUILD_ROOT